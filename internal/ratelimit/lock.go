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

func newLockToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.WithStack(err)
	}
	return hex.EncodeToString(buf), nil
}
