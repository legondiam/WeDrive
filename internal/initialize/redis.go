package initialize

import (
	"WeDrive/internal/config"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func RedisInit() (*redis.Client, error) {
	dbConn := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.GlobalConf.DB.Redis.Host, config.GlobalConf.DB.Redis.Port),
		Password: config.GlobalConf.DB.Redis.Password,
		DB:       0,
	})
	ctx := context.Background()
	_, err := dbConn.Ping(ctx).Result()
	if err != nil {
		dbConn.Close()
		return nil, errors.WithStack(err)
	}
	return dbConn, nil
}
