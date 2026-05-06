package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// SetJSON 序列化并设置到缓存
func SetJSON(ctx context.Context, client *redis.Client, key string, payload any, expire time.Duration) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := client.Set(ctx, key, raw, expire).Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetJSON 从缓存中获取并反序列化
func GetJSON(ctx context.Context, client *redis.Client, key string, target any) (bool, error) {
	raw, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

func Delete(ctx context.Context, client *redis.Client, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := client.Del(ctx, keys...).Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
