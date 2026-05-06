package repository

import (
	"WeDrive/internal/cache"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type UserCacheRepo struct {
	client *redis.Client
}

func NewUserCacheRepo(client *redis.Client) *UserCacheRepo {
	return &UserCacheRepo{client: client}
}

// SetRefreshToken 设置刷新令牌
func (c *UserCacheRepo) SetRefreshToken(ctx context.Context, userID uint, tokenID string, expire time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	err := c.client.Set(ctx, key, userID, expire).Err()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetRefreshToken 获取刷新令牌
func (c *UserCacheRepo) GetRefreshToken(ctx context.Context, userID uint, tokenID string) (bool, error) {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	value, err := c.client.Get(ctx, key).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return uint(value) == userID, nil
}

// DeleteRefreshToken 删除刷新令牌
func (c *UserCacheRepo) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *UserCacheRepo) SetUserInfo(ctx context.Context, user cache.UserInfo) error {
	return cache.SetJSON(ctx, c.client, cache.UserInfoKey(user.ID), user, cache.UserInfoTTL)
}

func (c *UserCacheRepo) GetUserInfo(ctx context.Context, userID uint) (*cache.UserInfo, bool, error) {
	var user cache.UserInfo
	ok, err := cache.GetJSON(ctx, c.client, cache.UserInfoKey(userID), &user)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &user, true, nil
}

func (c *UserCacheRepo) DeleteUserInfo(ctx context.Context, userID uint) error {
	return cache.Delete(ctx, c.client, cache.UserInfoKey(userID))
}
