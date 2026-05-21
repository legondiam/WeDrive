package service

import (
	"WeDrive/internal/cache"
	"WeDrive/internal/model"
	"WeDrive/internal/oss"
	"WeDrive/internal/ratelimit"
	"WeDrive/internal/repository"
	"WeDrive/pkg/logger"
	"WeDrive/pkg/utils/hash"
	"WeDrive/pkg/utils/jwts"
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var ErrFolderCannotShare = errors.New("文件夹不能分享")
var ErrShareInvalidKey = errors.New("密钥不正确")
var ErrShareExpired = errors.New("分享已过期")
var ErrShareNotFound = errors.New("分享不存在")

type ShareService struct {
	shareRepo   *repository.ShareRepo
	shareCache  *repository.ShareCacheRepo
	fileRepo    *repository.FileRepo
	rateLimiter *ratelimit.Limiter
	storage     *oss.Storage
	shareGroup  singleflight.Group
}

type shareDownloadResp struct {
	URL      string
	FileName string
}

func NewShareService(shareRepo *repository.ShareRepo, shareCache *repository.ShareCacheRepo, fileRepo *repository.FileRepo, rateLimiter *ratelimit.Limiter, storage *oss.Storage) *ShareService {
	return &ShareService{shareRepo: shareRepo, shareCache: shareCache, fileRepo: fileRepo, rateLimiter: rateLimiter, storage: storage}
}

// CreateShareFile 创建分享文件
func (s *ShareService) CreateShareFile(ctx context.Context, userID uint, userFileID uint, key string, expiretime *time.Time) (string, error) {
	// 检查文件是否存在
	file, err := s.fileRepo.GetFileByID(ctx, userFileID, userID)
	if err != nil {
		return "", errors.WithMessage(err, "获取文件失败")
	}
	if file.IsFolder {
		return "", ErrFolderCannotShare
	}
	//生成token
	token, err := jwts.GenerateToken(24)
	if err != nil {
		return "", errors.WithMessage(err, "生成token失败")
	}
	//生成keyhash
	keyhash := ""
	if key != "" {
		keyhash, err = hash.HashPassword(key)
		if err != nil {
			return "", errors.WithMessage(err, "生成keyhash失败")
		}
	}
	//创建分享文件
	shareFile := &model.ShareFile{
		UserID:     userID,
		UserFileID: userFileID,
		ShareToken: token,
		KeyHash:    keyhash,
		ExpiresAt:  expiretime,
	}
	err = s.shareRepo.CreateShareFile(ctx, shareFile)
	if err != nil {
		return "", errors.WithMessage(err, "创建分享文件失败")
	}
	if err := s.shareCache.SetShareToken(ctx, shareTokenCacheFromModel(shareFile)); err != nil {
		logger.S.Warnf("缓存分享文件失败:%v", err)
	}
	return token, nil
}

// GetShareDownloadURL 获取分享文件下载URL
func (s *ShareService) GetShareDownloadURL(ctx context.Context, token string, key string) (shareDownloadResp, error) {
	//获取分享文件
	shareFile, err := s.getShareFileByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shareDownloadResp{}, ErrShareNotFound
		}
		return shareDownloadResp{}, errors.WithMessage(err, "获取分享文件失败")
	}
	//校验是否过期
	if shareFile.ExpiresAt != nil && shareFile.ExpiresAt.Before(time.Now()) {
		return shareDownloadResp{}, errors.WithMessage(ErrShareExpired, "分享已过期")
	}
	//校验key
	if shareFile.KeyHash != "" {
		ok, err := hash.CheckPassword(key, shareFile.KeyHash)
		if err != nil {
			return shareDownloadResp{}, errors.WithMessage(err, "校验key失败")
		}
		if !ok {
			return shareDownloadResp{}, errors.WithMessage(ErrShareInvalidKey, "密钥不正确")
		}
	}
	//获取文件
	fileStore, fileName, err := s.fileRepo.GetFileStoreByID(ctx, shareFile.UserFileID, shareFile.UserID)
	if err != nil {
		return shareDownloadResp{}, errors.WithMessage(err, "获取文件失败")
	}
	//获取下载URL
	url, err := s.storage.DownloadFile(ctx, fileStore.FileAddr, fileName, 15*time.Minute, "free")
	if err != nil {
		return shareDownloadResp{}, errors.WithMessage(err, "获取下载URL失败")
	}
	return shareDownloadResp{URL: url, FileName: fileName}, nil
}

// getShareFileByToken 通过 token 获取分享记录并按需重建缓存。
func (s *ShareService) getShareFileByToken(ctx context.Context, token string) (*model.ShareFile, error) {
	cachedShare, ok, err := s.shareCache.GetShareToken(ctx, token)
	if err != nil {
		logger.S.Warnf("读取分享缓存失败:%v", err)
		return s.getShareFileByTokenSingleflight(ctx, token)
	}
	if ok {
		return shareFileFromCache(cachedShare), nil
	}

	//缓存未命中，获取锁
	lockKey := "lock:cache:share:token:" + token
	lockToken, locked, err := s.rateLimiter.TryLock(ctx, lockKey, cacheRebuildLockTTL)
	if err != nil {
		logger.S.Warnf("获取分享缓存重建锁失败:%v", err)
		return s.getShareFileByTokenSingleflight(ctx, token)
	}
	//获取到锁，重建缓存
	if locked {
		defer func() {
			_ = s.rateLimiter.Unlock(context.Background(), lockKey, lockToken)
		}()
		//再次查缓存
		cachedShare, ok, err = s.shareCache.GetShareToken(ctx, token)
		if err != nil {
			logger.S.Warnf("读取分享缓存失败:%v", err)
			return s.getShareFileByTokenSingleflight(ctx, token)
		}
		if ok {
			return shareFileFromCache(cachedShare), nil
		}
		shareFile, err := s.shareRepo.GetShareFile(ctx, token)
		if err != nil {
			return nil, err
		}
		if err := s.shareCache.SetShareToken(ctx, shareTokenCacheFromModel(shareFile)); err != nil {
			logger.S.Warnf("缓存分享文件失败:%v", err)
		}
		return shareFile, nil
	}

	//获取不到锁，重试
	for i := 0; i < cacheRebuildRetry; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cacheRebuildWait):
		}
		cachedShare, ok, err = s.shareCache.GetShareToken(ctx, token)
		if err != nil {
			logger.S.Warnf("读取分享缓存失败:%v", err)
			return s.getShareFileByTokenSingleflight(ctx, token)
		}
		if ok {
			return shareFileFromCache(cachedShare), nil
		}
	}

	logger.S.Warnf("等待分享缓存重建超时, token: %s", token)
	return s.getShareFileByTokenSingleflight(ctx, token)
}

// getShareFileByTokenSingleflight 合并同一 token 的数据库查询。
func (s *ShareService) getShareFileByTokenSingleflight(ctx context.Context, token string) (*model.ShareFile, error) {
	value, err, _ := s.shareGroup.Do(token, func() (any, error) {
		return s.shareRepo.GetShareFile(ctx, token)
	})
	if err != nil {
		return nil, err
	}
	shareFile, ok := value.(*model.ShareFile)
	if !ok {
		return nil, errors.New("share singleflight result invalid")
	}
	return shareFile, nil
}

// shareTokenCacheFromModel 将分享模型转换为缓存结构。
func shareTokenCacheFromModel(shareFile *model.ShareFile) cache.ShareToken {
	return cache.ShareToken{
		ID:         shareFile.ID,
		UserID:     shareFile.UserID,
		UserFileID: shareFile.UserFileID,
		ShareToken: shareFile.ShareToken,
		KeyHash:    shareFile.KeyHash,
		ExpiresAt:  shareFile.ExpiresAt,
	}
}

// shareFileFromCache 将分享缓存转换为模型。
func shareFileFromCache(item *cache.ShareToken) *model.ShareFile {
	return &model.ShareFile{
		UserID:     item.UserID,
		UserFileID: item.UserFileID,
		ShareToken: item.ShareToken,
		KeyHash:    item.KeyHash,
		ExpiresAt:  item.ExpiresAt,
	}
}
