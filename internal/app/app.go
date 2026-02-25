package app

import (
	"WeDrive/internal/initialize"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// 全局依赖容器

func Init() (error, *gin.Engine) {
	db, err := initialize.MysqlInit()
	if err != nil {
		return errors.WithMessage(err, "mysql初始化失败"), nil
	}
	redis, err := initialize.RedisInit()
	if err != nil {
		return errors.WithMessage(err, "redis初始化失败"), nil
	}
	minio, err := initialize.MinioInit()
	if err != nil {
		return errors.WithMessage(err, "minio初始化失败"), nil
	}
	engine := BuildApp(db, redis, minio)
	return nil, engine
}
