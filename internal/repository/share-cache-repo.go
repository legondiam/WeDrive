package repository

import (
	"WeDrive/internal/cache"
	"WeDrive/internal/cacheguard"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ShareCacheRepo struct {
	client *redis.Client
	guard  *cacheguard.RedisGuard
}

func NewShareCacheRepo(client *redis.Client, guard *cacheguard.RedisGuard) *ShareCacheRepo {
	return &ShareCacheRepo{client: client, guard: guard}
}

// SetShareToken 缓存分享 token 信息。
func (r *ShareCacheRepo) SetShareToken(ctx context.Context, item cache.ShareToken) error {
	expire := cache.ShareTokenTTL
	if item.ExpiresAt != nil {
		remaining := time.Until(*item.ExpiresAt)
		if remaining <= 0 {
			return nil
		}
		expire = cache.JitterTTLWithin(expire, remaining)
	} else {
		expire = cache.JitterTTL(expire)
	}
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.SetJSON(ctx, r.client, cache.ShareTokenKey(item.ShareToken), item, expire)
	})
}

// GetShareToken 读取分享 token 缓存。
func (r *ShareCacheRepo) GetShareToken(ctx context.Context, token string) (*cache.ShareToken, bool, error) {
	var item cache.ShareToken
	ok, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (bool, error) {
		return cache.GetJSON(ctx, r.client, cache.ShareTokenKey(token), &item)
	})
	if err != nil || !ok {
		return nil, ok, err
	}
	return &item, true, nil
}

// DeleteShareToken 删除分享 token 缓存。
func (r *ShareCacheRepo) DeleteShareToken(ctx context.Context, token string) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return cache.Delete(ctx, r.client, cache.ShareTokenKey(token))
	})
}
