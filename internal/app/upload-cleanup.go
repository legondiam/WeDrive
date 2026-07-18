package app

import (
	"WeDrive/internal/cacheguard"
	"WeDrive/internal/config"
	"WeDrive/internal/mq"
	"WeDrive/internal/oss"
	"WeDrive/internal/ratelimit"
	"WeDrive/internal/repository"
	"WeDrive/internal/service"
	"WeDrive/pkg/logger"
	"context"
	"time"

	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
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
func startBackgroundJobs(db *gorm.DB, redisClient *redis.Client, minioClient *minio.Client, mqConn *amqp.Connection) {
	redisGuard := cacheguard.NewRedisGuard()
	bloomRepo := repository.NewBloomRepo(redisClient, redisGuard)
	fileRepo := repository.NewFileRepo(db)
	shareRepo := repository.NewShareRepo(db)
	startCacheInvalidationConsumer(redisClient, redisGuard, mqConn)
	startBloomRepairConsumer(bloomRepo, mqConn)
	startUploadVerificationConsumer(db, redisClient, redisGuard, bloomRepo, minioClient, mqConn)
	startExpiredUploadCleanup(db, redisClient, redisGuard, bloomRepo, minioClient)
	startBloomWarmup(fileRepo, shareRepo, bloomRepo)
}

func startCacheInvalidationConsumer(redisClient *redis.Client, redisGuard *cacheguard.RedisGuard, mqConn *amqp.Connection) {
	userCacheRepo := repository.NewUserCacheRepo(redisClient, redisGuard)
	fileCacheRepo := repository.NewFileCacheRepo(redisClient, redisGuard)
	if err := mq.StartCacheInvalidationConsumer(mqConn, userCacheRepo, fileCacheRepo); err != nil {
		logger.S.Errorf("缓存失效消费者启动失败: %+v", err)
		return
	}
	logger.S.Info("缓存失效消费者已启动")
}

func startBloomRepairConsumer(bloomRepo *repository.BloomRepo, mqConn *amqp.Connection) {
	if err := mq.StartBloomRepairConsumer(mqConn, bloomRepo); err != nil {
		logger.S.Errorf("布隆补偿消费者启动失败: %+v", err)
		return
	}
	logger.S.Info("布隆补偿消费者已启动")
}

func startUploadVerificationConsumer(db *gorm.DB, redisClient *redis.Client, redisGuard *cacheguard.RedisGuard, bloomRepo *repository.BloomRepo, minioClient *minio.Client, mqConn *amqp.Connection) {
	fileRepo := repository.NewFileRepo(db)
	fileCacheRepo := repository.NewFileCacheRepo(redisClient, redisGuard)
	uploadCacheRepo := repository.NewUploadCacheRepo(redisClient, redisGuard)
	rateLimiter := ratelimit.NewLimiter(redisClient, redisGuard)
	userRepo := repository.NewUserRepo(db)
	userCacheRepo := repository.NewUserCacheRepo(redisClient, redisGuard)
	storage := oss.NewStorage(minioClient)
	cachePublisher := mq.NewCacheInvalidationPublisher(mqConn)
	uploadVerifier := mq.NewUploadVerificationPublisher(mqConn)
	bloomPublisher := mq.NewBloomRepairPublisher(mqConn)
	fileService := service.NewFileService(fileRepo, fileCacheRepo, uploadCacheRepo, rateLimiter, userRepo, userCacheRepo, bloomRepo, storage, db, cachePublisher, uploadVerifier, bloomPublisher)
	if err := mq.StartUploadVerificationConsumer(mqConn, fileService.VerifyChunkUpload); err != nil {
		logger.S.Errorf("上传校验消费者启动失败: %+v", err)
		return
	}
	logger.S.Info("上传校验消费者已启动")
}

// startExpiredUploadCleanup 启动僵尸分块定时清理任务。
func startExpiredUploadCleanup(db *gorm.DB, redisClient *redis.Client, redisGuard *cacheguard.RedisGuard, bloomRepo *repository.BloomRepo, minioClient *minio.Client) {
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
	fileCacheRepo := repository.NewFileCacheRepo(redisClient, redisGuard)
	uploadCacheRepo := repository.NewUploadCacheRepo(redisClient, redisGuard)
	rateLimiter := ratelimit.NewLimiter(redisClient, redisGuard)
	userRepo := repository.NewUserRepo(db)
	userCacheRepo := repository.NewUserCacheRepo(redisClient, redisGuard)
	storage := oss.NewStorage(minioClient)
	fileService := service.NewFileService(fileRepo, fileCacheRepo, uploadCacheRepo, rateLimiter, userRepo, userCacheRepo, bloomRepo, storage, db, nil, nil, nil)

	go func() {
		// 服务启动后先执行一轮，减少历史残留会话堆积时间。
		runUploadCleanup(fileService, expireAfter, batchSize)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runUploadCleanup(fileService, expireAfter, batchSize)
		}
	}()

	logger.S.Infof("僵尸分块清理任务已启动, interval: %s, expireAfter: %s, batchSize: %d", interval, expireAfter, batchSize)
}

// runUploadCleanup 执行单轮清理
func runUploadCleanup(fileService *service.FileService, expireAfter time.Duration, batchSize int) {
	expireBefore := time.Now().Add(-expireAfter)
	cleaned, err := fileService.CleanupExpiredUploadSessions(context.Background(), expireBefore, batchSize)
	if err != nil {
		logger.S.Errorf("执行僵尸分块清理失败: %+v", err)
		return
	}
	if cleaned > 0 {
		logger.S.Infof("本轮僵尸分块清理完成, cleaned: %d", cleaned)
	}
}
