package repository

import (
	"WeDrive/internal/cache"
	"WeDrive/internal/cacheguard"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type UserCacheRepo struct {
	client *redis.Client
	guard  *cacheguard.RedisGuard
}

func NewUserCacheRepo(client *redis.Client, guard *cacheguard.RedisGuard) *UserCacheRepo {
	return &UserCacheRepo{client: client, guard: guard}
}

// SetRefreshToken 缓存刷新令牌。
func (c *UserCacheRepo) SetRefreshToken(ctx context.Context, userID uint, tokenID string, expire time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	err := c.guard.Do(ctx, func(ctx context.Context) error {
		return c.client.Set(ctx, key, userID, expire).Err()
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetRefreshToken 读取刷新令牌并校验归属用户。
func (c *UserCacheRepo) GetRefreshToken(ctx context.Context, userID uint, tokenID string) (bool, error) {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	value, err := cacheguard.DoResult(c.guard, ctx, func(ctx context.Context) (uint64, error) {
		return c.client.Get(ctx, key).Uint64()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return uint(value) == userID, nil
}

// DeleteRefreshToken 删除刷新令牌。
func (c *UserCacheRepo) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	err := c.guard.Do(ctx, func(ctx context.Context) error {
		return c.client.Del(ctx, key).Err()
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SetUserInfo 缓存用户信息。
func (c *UserCacheRepo) SetUserInfo(ctx context.Context, user cache.UserInfo) error {
	vipExpireAt := ""
	if user.VipExpireAt != nil {
		vipExpireAt = user.VipExpireAt.Format(time.RFC3339Nano)
	}
	key := cache.UserInfoKey(user.ID)
	if err := c.guard.Do(ctx, func(ctx context.Context) error {
		pipe := c.client.TxPipeline()
		pipe.HSet(ctx, key, map[string]any{
			"id":            strconv.FormatUint(uint64(user.ID), 10),
			"username":      user.Username,
			"total_space":   strconv.FormatInt(user.TotalSpace, 10),
			"used_space":    strconv.FormatInt(user.UsedSpace, 10),
			"member_level":  strconv.FormatInt(int64(user.MemberLevel), 10),
			"vip_expire_at": vipExpireAt,
		})
		pipe.Expire(ctx, key, cache.JitterTTL(cache.UserInfoTTL))
		_, err := pipe.Exec(ctx)
		return err
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetUserInfo 从缓存读取用户信息。
func (c *UserCacheRepo) GetUserInfo(ctx context.Context, userID uint) (*cache.UserInfo, bool, error) {
	values, err := cacheguard.DoResult(c.guard, ctx, func(ctx context.Context) (map[string]string, error) {
		return c.client.HGetAll(ctx, cache.UserInfoKey(userID)).Result()
	})
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	if len(values) == 0 {
		return nil, false, nil
	}
	id, err := strconv.ParseUint(values["id"], 10, 64)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	totalSpace, err := strconv.ParseInt(values["total_space"], 10, 64)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	usedSpace, err := strconv.ParseInt(values["used_space"], 10, 64)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	memberLevel, err := strconv.ParseInt(values["member_level"], 10, 8)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	var vipExpireAt *time.Time
	if values["vip_expire_at"] != "" {
		parsed, err := time.Parse(time.RFC3339Nano, values["vip_expire_at"])
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
		vipExpireAt = &parsed
	}
	return &cache.UserInfo{
		ID:          uint(id),
		Username:    values["username"],
		TotalSpace:  totalSpace,
		UsedSpace:   usedSpace,
		MemberLevel: int8(memberLevel),
		VipExpireAt: vipExpireAt,
	}, true, nil
}

// DeleteUserInfo 删除用户信息缓存。
func (c *UserCacheRepo) DeleteUserInfo(ctx context.Context, userID uint) error {
	return c.guard.Do(ctx, func(ctx context.Context) error {
		return cache.Delete(ctx, c.client, cache.UserInfoKey(userID))
	})
}
