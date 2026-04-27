package app

import (
	"WeDrive/internal/config"
	"WeDrive/internal/oss"
	"WeDrive/internal/repository"
	"WeDrive/internal/service"
	"WeDrive/pkg/logger"
	"context"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	//默认清理配置
	defaultUploadCleanupInterval    = 10 * time.Minute
	defaultUploadCleanupExpireAfter = 24 * time.Hour
	defaultUploadCleanupBatchSize   = 100
)

// startBackgroundJobs 启动后台任务
func startBackgroundJobs(db *gorm.DB, redisClient *redis.Client, minioClient *minio.Client) {
	startUploadSessionCleanup(db, redisClient, minioClient)
}

// startUploadSessionCleanup 启动上传会话定时清理任务。
func startUploadSessionCleanup(db *gorm.DB, redisClient *redis.Client, minioClient *minio.Client) {
	cleanupConf := config.GlobalConf.UploadCleanup
	if !cleanupConf.Enabled {
		return
	}
	//确保配置有效
	interval := cleanupConf.Interval
	if interval <= 0 {
		interval = defaultUploadCleanupInterval
	}
	expireAfter := cleanupConf.ExpireAfter
	if expireAfter <= 0 {
		expireAfter = defaultUploadCleanupExpireAfter
	}
	batchSize := cleanupConf.BatchSize
	if batchSize <= 0 {
		batchSize = defaultUploadCleanupBatchSize
	}

	fileRepo := repository.NewFileRepo(db)
	uploadCacheRepo := repository.NewUploadCacheRepo(redisClient)
	userRepo := repository.NewUserRepo(db)
	storage := oss.NewStorage(minioClient)
	fileService := service.NewFileService(fileRepo, uploadCacheRepo, userRepo, storage, db)

	go func() {
		// 服务启动后先执行一轮，减少历史残留会话堆积时间。
		runUploadCleanup(fileService, expireAfter, batchSize)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runUploadCleanup(fileService, expireAfter, batchSize)
		}
	}()

	logger.S.Infof("上传会话清理任务已启动, interval: %s, expireAfter: %s, batchSize: %d", interval, expireAfter, batchSize)
}

// runUploadCleanup 执行单轮上传会话清理
func runUploadCleanup(fileService *service.FileService, expireAfter time.Duration, batchSize int) {
	expireBefore := time.Now().Add(-expireAfter)
	cleaned, err := fileService.CleanupStaleUploadSessions(context.Background(), expireBefore, batchSize)
	if err != nil {
		logger.S.Errorf("执行上传会话清理失败: %+v", err)
		return
	}
	if cleaned > 0 {
		logger.S.Infof("本轮上传会话清理完成, cleaned: %d", cleaned)
	}
}
