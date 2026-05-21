//go:build wireinject
// +build wireinject

package app

import (
	"WeDrive/internal/api"
	"WeDrive/internal/cacheguard"
	"WeDrive/internal/mq"
	"WeDrive/internal/oss"
	"WeDrive/internal/ratelimit"
	"WeDrive/internal/repository"
	"WeDrive/internal/router"
	"WeDrive/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func BuildApp(db *gorm.DB, redis *redis.Client, minio *minio.Client, mqConn *amqp.Connection) *gin.Engine {
	// wire 会自动分析依赖顺序：db -> Repo -> Service -> Handler -> Router
	wire.Build(
		mq.NewCacheInvalidationPublisher,
		mq.NewUploadVerificationPublisher,
		cacheguard.NewRedisGuard,
		repository.NewUserRepo,
		repository.NewUserCacheRepo,
		repository.NewFileRepo,
		repository.NewFileCacheRepo,
		repository.NewUploadCacheRepo,
		ratelimit.NewLimiter,
		repository.NewShareRepo,
		repository.NewShareCacheRepo,
		oss.NewStorage,
		service.NewUserService,
		service.NewFileService,
		service.NewShareService,
		api.NewUserHandler,
		api.NewFileHandler,
		api.NewShareHandler,
		router.NewRouter,
	)
	return nil
}
