package app

import (
	"WeDrive/internal/initialize"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	Engine *gin.Engine
	db     *gorm.DB
	redis  *redis.Client
	minio  *minio.Client
	mq     *amqp.Connection
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
	mqConn, err := initialize.RabbitMQInit()
	if err != nil {
		return nil, errors.WithMessage(err, "rabbitmq初始化失败")
	}
	engine := BuildApp(db, redisClient, minioClient, mqConn)
	return &App{
		Engine: engine,
		db:     db,
		redis:  redisClient,
		minio:  minioClient,
		mq:     mqConn,
	}, nil
}

func (a *App) StartBackgroundJobs() {
	if a == nil {
		return
	}
	startBackgroundJobs(a.db, a.redis, a.minio, a.mq)
}

func (a *App) Close() {
	if a == nil {
		return
	}
	if a.mq != nil {
		_ = a.mq.Close()
	}
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.db != nil {
		if db, err := a.db.DB(); err == nil {
			_ = db.Close()
		}
	}
}
