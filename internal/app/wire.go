//go:build wireinject
// +build wireinject

package app

import (
	"WeDrive/internal/api"
	"WeDrive/internal/oss"
	"WeDrive/internal/repository"
	"WeDrive/internal/router"
	"WeDrive/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func BuildApp(db *gorm.DB, redis *redis.Client, minio *minio.Client) *gin.Engine {
	// wire 会自动分析依赖顺序：db -> Repo -> Service -> Handler -> Router
	wire.Build(
		repository.NewUserRepo,
		repository.NewUserCacheRepo,
		repository.NewFileRepo,
		repository.NewUploadCacheRepo,
		repository.NewShareRepo,
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
