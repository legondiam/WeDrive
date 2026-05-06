package repository

import (
	"WeDrive/internal/cache"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ShareCacheRepo struct {
	client *redis.Client
}

func NewShareCacheRepo(client *redis.Client) *ShareCacheRepo {
	return &ShareCacheRepo{client: client}
}

func (r *ShareCacheRepo) SetShareToken(ctx context.Context, item cache.ShareToken) error {
	expire := cache.ShareTokenTTL
	if item.ExpiresAt != nil {
		remaining := time.Until(*item.ExpiresAt)
		if remaining <= 0 {
			return nil
		}
		if remaining < expire {
			expire = remaining
		}
	}
	return cache.SetJSON(ctx, r.client, cache.ShareTokenKey(item.ShareToken), item, expire)
}

func (r *ShareCacheRepo) GetShareToken(ctx context.Context, token string) (*cache.ShareToken, bool, error) {
	var item cache.ShareToken
	ok, err := cache.GetJSON(ctx, r.client, cache.ShareTokenKey(token), &item)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &item, true, nil
}

func (r *ShareCacheRepo) DeleteShareToken(ctx context.Context, token string) error {
	return cache.Delete(ctx, r.client, cache.ShareTokenKey(token))
}
