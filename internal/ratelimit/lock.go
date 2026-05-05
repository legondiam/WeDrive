package ratelimit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/pkg/errors"
)

const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

const refreshLockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`

// TryLock 获取分布式锁
func (l *Limiter) TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	if ttl <= 0 {
		return "", false, errors.New("lock ttl invalid")
	}
	token, err := newLockToken()
	if err != nil {
		return "", false, err
	}
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return token, ok, nil
}

// Unlock 在token匹配时释放分布式锁
func (l *Limiter) Unlock(ctx context.Context, key, token string) error {
	if key == "" || token == "" {
		return errors.New("lock key or token invalid")
	}
	if _, err := l.client.Eval(ctx, unlockScript, []string{key}, token).Result(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RefreshLock 在token匹配时延长锁TTL
func (l *Limiter) RefreshLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	if key == "" || token == "" || ttl <= 0 {
		return false, errors.New("lock refresh config invalid")
	}
	result, err := l.client.Eval(ctx, refreshLockScript, []string{key}, token, ttl.Milliseconds()).Int()
	if err != nil {
		return false, errors.WithStack(err)
	}
	return result == 1, nil
}

// AutoRefreshLock 定时续期锁，直到ctx结束、续期失败或锁不再属于当前token
func (l *Limiter) AutoRefreshLock(ctx context.Context, key, token string, ttl time.Duration, onError func(error)) {
	interval := ttl / 3
	if interval <= 0 {
		interval = ttl
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ok, err := l.RefreshLock(ctx, key, token, ttl)
				if err != nil {
					if onError != nil {
						onError(err)
					}
					return
				}
				if !ok {
					return
				}
			}
		}
	}()
}

func newLockToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.WithStack(err)
	}
	return hex.EncodeToString(buf), nil
}
