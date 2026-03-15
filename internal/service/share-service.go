package service

import (
	"WeDrive/internal/model"
	"WeDrive/internal/oss"
	"WeDrive/internal/repository"
	"WeDrive/pkg/utils/hash"
	"WeDrive/pkg/utils/jwts"
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var ErrFolderCannotShare = errors.New("文件夹不能分享")
var ErrShareInvalidKey = errors.New("密钥不正确")
var ErrShareExpired = errors.New("分享已过期")
var ErrShareNotFound = errors.New("分享不存在")

type ShareService struct {
	shareRepo *repository.ShareRepo
	fileRepo  *repository.FileRepo
	storage   *oss.Storage
}

type shareDownloadResp struct {
	URL      string
	FileName string
}

func NewShareService(shareRepo *repository.ShareRepo, fileRepo *repository.FileRepo, storage *oss.Storage) *ShareService {
	return &ShareService{shareRepo: shareRepo, fileRepo: fileRepo, storage: storage}
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
	return token, nil
}

// GetShareDownloadURL 获取分享文件下载URL
func (s *ShareService) GetShareDownloadURL(ctx context.Context, token string, key string) (shareDownloadResp, error) {
	//获取分享文件
	shareFile, err := s.shareRepo.GetShareFile(ctx, token)
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
