package repository

import (
	"WeDrive/internal/cache"
	"WeDrive/internal/cacheguard"
	"WeDrive/internal/model"
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type FileCacheRepo struct {
	client *redis.Client
	guard  *cacheguard.RedisGuard
}

func NewFileCacheRepo(client *redis.Client, guard *cacheguard.RedisGuard) *FileCacheRepo {
	return &FileCacheRepo{client: client, guard: guard}
}

// SetUserFileList 缓存用户指定目录下的文件列表
func (r *FileCacheRepo) SetUserFileList(ctx context.Context, userID uint, parentID uint, list []cache.FileListItem) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.SetJSON(ctx, r.client, cache.UserFileListKey(userID, parentID), list, cache.JitterTTL(cache.FileListTTL))
	})
}

// GetUserFileList 获取用户指定目录下的文件列表缓存
func (r *FileCacheRepo) GetUserFileList(ctx context.Context, userID uint, parentID uint) ([]cache.FileListItem, bool, error) {
	var list []cache.FileListItem
	ok, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (bool, error) {
		return cache.GetJSON(ctx, r.client, cache.UserFileListKey(userID, parentID), &list)
	})
	return list, ok, err
}

// DeleteUserFileList 删除用户指定目录下的文件列表缓存
func (r *FileCacheRepo) DeleteUserFileList(ctx context.Context, userID uint, parentID uint) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.Delete(ctx, r.client, cache.UserFileListKey(userID, parentID))
	})
}

// SetRecycleBinList 缓存用户回收站文件列表
func (r *FileCacheRepo) SetRecycleBinList(ctx context.Context, userID uint, list []cache.RecycleFileListItem) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.SetJSON(ctx, r.client, cache.RecycleBinListKey(userID), list, cache.JitterTTL(cache.FileListTTL))
	})
}

// GetRecycleBinList 获取用户回收站文件列表缓存
func (r *FileCacheRepo) GetRecycleBinList(ctx context.Context, userID uint) ([]cache.RecycleFileListItem, bool, error) {
	var list []cache.RecycleFileListItem
	ok, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (bool, error) {
		return cache.GetJSON(ctx, r.client, cache.RecycleBinListKey(userID), &list)
	})
	return list, ok, err
}

// DeleteRecycleBinList 删除用户回收站文件列表缓存
func (r *FileCacheRepo) DeleteRecycleBinList(ctx context.Context, userID uint) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.Delete(ctx, r.client, cache.RecycleBinListKey(userID))
	})
}

// SetDownloadFileMeta 缓存生成下载URL所需的文件池元数据
func (r *FileCacheRepo) SetDownloadFileMeta(ctx context.Context, userID uint, userFileID uint, meta cache.DownloadFileMeta) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.SetJSON(ctx, r.client, cache.DownloadFileMetaKey(userID, userFileID), meta, cache.JitterTTL(cache.FileMetaTTL))
	})
}

// GetDownloadFileMeta 获取生成下载URL所需的文件池元数据缓存
func (r *FileCacheRepo) GetDownloadFileMeta(ctx context.Context, userID uint, userFileID uint) (*cache.DownloadFileMeta, bool, error) {
	var meta cache.DownloadFileMeta
	ok, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (bool, error) {
		return cache.GetJSON(ctx, r.client, cache.DownloadFileMetaKey(userID, userFileID), &meta)
	})
	if err != nil || !ok {
		return nil, ok, err
	}
	return &meta, true, nil
}

// DeleteDownloadFileMeta 删除生成下载URL所需的文件池元数据缓存
func (r *FileCacheRepo) DeleteDownloadFileMeta(ctx context.Context, userID uint, userFileID uint) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.Delete(ctx, r.client, cache.DownloadFileMetaKey(userID, userFileID))
	})
}

// SetFileIdentity 缓存文件池身份信息
func (r *FileCacheRepo) SetFileIdentity(ctx context.Context, file *model.FileStore) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.SetJSON(ctx, r.client, cache.FileIdentityKey(file.HashType, file.FileHash), fileIdentityFromModel(file), cache.JitterTTL(cache.FileIdentityTTL))
	})
}

// GetFileIdentity 根据哈希身份获取文件池身份缓存
func (r *FileCacheRepo) GetFileIdentity(ctx context.Context, hashType string, fileHash string) (*model.FileStore, bool, error) {
	var item cache.FileIdentity
	ok, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (bool, error) {
		return cache.GetJSON(ctx, r.client, cache.FileIdentityKey(hashType, fileHash), &item)
	})
	if err != nil || !ok {
		return nil, ok, err
	}
	return fileStoreFromIdentity(item), true, nil
}

// DeleteFileIdentity 删除文件池身份缓存
func (r *FileCacheRepo) DeleteFileIdentity(ctx context.Context, hashType string, fileHash string) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.Delete(ctx, r.client, cache.FileIdentityKey(hashType, fileHash))
	})
}

// SetFileSample 缓存抽样哈希是否命中文件池
func (r *FileCacheRepo) SetFileSample(ctx context.Context, fileSize int64, headHash string, midHash string, tailHash string, exists bool) error {
	value := "0"
	if exists {
		value = "1"
	}
	if err := r.guard.Do(ctx, func(ctx context.Context) error {
		return r.client.Set(ctx, cache.FileSampleKey(fileSize, headHash, midHash, tailHash), value, cache.JitterTTL(cache.FileSampleTTL)).Err()
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetFileSample 根据抽样哈希获取文件池命中缓存
func (r *FileCacheRepo) GetFileSample(ctx context.Context, fileSize int64, headHash string, midHash string, tailHash string) (bool, bool, error) {
	value, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (string, error) {
		return r.client.Get(ctx, cache.FileSampleKey(fileSize, headHash, midHash, tailHash)).Result()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, false, nil
		}
		return false, false, errors.WithStack(err)
	}
	return value == "1", true, nil
}

// fileIdentityFromModel 将文件池模型转换为缓存结构
func fileIdentityFromModel(file *model.FileStore) cache.FileIdentity {
	return cache.FileIdentity{
		ID:        file.ID,
		HashType:  file.HashType,
		FileHash:  file.FileHash,
		FileName:  file.FileName,
		FileSize:  file.FileSize,
		FileAddr:  file.FileAddr,
		HeadHash:  file.HeadHash,
		MidHash:   file.MidHash,
		TailHash:  file.TailHash,
		CreatedAt: file.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: file.UpdatedAt.Format(time.RFC3339Nano),
	}
}

// fileStoreFromIdentity 将文件池身份缓存还原为文件池模型
func fileStoreFromIdentity(item cache.FileIdentity) *model.FileStore {
	createdAt, err := time.Parse(time.RFC3339Nano, item.CreatedAt)
	if err != nil {
		createdAt = time.Time{}
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, item.UpdatedAt)
	if err != nil {
		updatedAt = time.Time{}
	}
	return &model.FileStore{
		Model: gorm.Model{
			ID:        item.ID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		HashType: item.HashType,
		FileHash: item.FileHash,
		FileName: item.FileName,
		FileSize: item.FileSize,
		FileAddr: item.FileAddr,
		HeadHash: item.HeadHash,
		MidHash:  item.MidHash,
		TailHash: item.TailHash,
	}
}
