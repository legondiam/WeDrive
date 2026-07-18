package app

import (
	"WeDrive/internal/cache"
	"WeDrive/internal/config"
	"WeDrive/internal/repository"
	"WeDrive/pkg/logger"
	"context"
	"time"

	"github.com/pkg/errors"
)

const defaultBloomWarmupBatchSize = 1000

// startBloomWarmup 启动布隆过滤器启动构建任务。
func startBloomWarmup(fileRepo *repository.FileRepo, shareRepo *repository.ShareRepo, bloomRepo *repository.BloomRepo) {
	if !config.GlobalConf.Bloom.Enabled {
		return
	}
	batchSize := config.GlobalConf.Bloom.WarmupBatchSize
	if batchSize <= 0 {
		batchSize = defaultBloomWarmupBatchSize
	}
	go func() {
		ctx := context.Background()
		if err := warmupFileIdentityBloom(ctx, fileRepo, bloomRepo, batchSize); err != nil {
			logger.S.Warnf("文件身份布隆过滤器构建失败:%v", err)
		}
		if err := warmupShareTokenBloom(ctx, shareRepo, bloomRepo, batchSize); err != nil {
			logger.S.Warnf("分享 token 布隆过滤器构建失败:%v", err)
		}
	}()
}

// warmupFileIdentityBloom 构建文件身份布隆过滤器。
func warmupFileIdentityBloom(ctx context.Context, fileRepo *repository.FileRepo, bloomRepo *repository.BloomRepo, batchSize int) error {
	if err := bloomRepo.Clear(ctx, cache.BloomFileIdentity); err != nil {
		return errors.WithMessage(err, "清空文件身份布隆过滤器失败")
	}
	if err := bloomRepo.SetReady(ctx, cache.BloomFileIdentity, false); err != nil {
		return errors.WithMessage(err, "关闭文件身份布隆过滤器失败")
	}
	var lastID uint
	for {
		files, err := fileRepo.ListFileIdentityAfterID(ctx, lastID, batchSize)
		if err != nil {
			return errors.WithMessage(err, "查询文件身份列表失败")
		}
		if len(files) == 0 {
			break
		}
		items := make([]string, 0, len(files))
		for _, file := range files {
			items = append(items, cache.FileIdentityBloomItem(file.HashType, file.FileHash))
			lastID = file.ID
		}
		if err := bloomRepo.AddMany(ctx, cache.BloomFileIdentity, items); err != nil {
			return errors.WithMessage(err, "批量写入文件身份布隆过滤器失败")
		}
	}
	if err := bloomRepo.SetReady(ctx, cache.BloomFileIdentity, true); err != nil {
		return errors.WithMessage(err, "启用文件身份布隆过滤器失败")
	}
	logger.S.Info("文件身份布隆过滤器构建完成")
	return nil
}

// warmupShareTokenBloom 构建分享 token 布隆过滤器。
func warmupShareTokenBloom(ctx context.Context, shareRepo *repository.ShareRepo, bloomRepo *repository.BloomRepo, batchSize int) error {
	if err := bloomRepo.Clear(ctx, cache.BloomShareToken); err != nil {
		return errors.WithMessage(err, "清空分享 token 布隆过滤器失败")
	}
	if err := bloomRepo.SetReady(ctx, cache.BloomShareToken, false); err != nil {
		return errors.WithMessage(err, "关闭分享 token 布隆过滤器失败")
	}
	var lastID uint
	now := time.Now()
	for {
		shares, err := shareRepo.ListActiveShareTokenAfterID(ctx, lastID, batchSize, now)
		if err != nil {
			return errors.WithMessage(err, "查询有效分享 token 列表失败")
		}
		if len(shares) == 0 {
			break
		}
		items := make([]string, 0, len(shares))
		for _, share := range shares {
			items = append(items, share.ShareToken)
			lastID = share.ID
		}
		if err := bloomRepo.AddMany(ctx, cache.BloomShareToken, items); err != nil {
			return errors.WithMessage(err, "批量写入分享 token 布隆过滤器失败")
		}
	}
	if err := bloomRepo.SetReady(ctx, cache.BloomShareToken, true); err != nil {
		return errors.WithMessage(err, "启用分享 token 布隆过滤器失败")
	}
	logger.S.Info("分享 token 布隆过滤器构建完成")
	return nil
}
