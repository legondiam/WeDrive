package app

import (
	"WeDrive/internal/initialize"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	Engine *gin.Engine
	db     *gorm.DB
	redis  *redis.Client
	minio  *minio.Client
}

func Init() (*App, error) {
	db, err := initialize.MysqlInit()
	if err != nil {
		return nil, errors.WithMessage(err, "mysql初始化失败")
	}
	redisClient, err := initialize.RedisInit()
	if err != nil {
		return nil, errors.WithMessage(err, "redis初始化失败")
	}
	minioClient, err := initialize.MinioInit()
	if err != nil {
		return nil, errors.WithMessage(err, "minio初始化失败")
	}
	engine := BuildApp(db, redisClient, minioClient)
	return &App{
		Engine: engine,
		db:     db,
		redis:  redisClient,
		minio:  minioClient,
	}, nil
}

func (a *App) StartBackgroundJobs() {
	if a == nil {
		return
	}
	startBackgroundJobs(a.db, a.redis, a.minio)
}
