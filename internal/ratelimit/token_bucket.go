package ratelimit

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const tokenBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

local redis_time = redis.call("TIME")
local now = redis_time[1] * 1000 + math.floor(redis_time[2] / 1000)

local state = redis.call("HMGET", key, "tokens", "updated_at")
local tokens = tonumber(state[1])
local updated_at = tonumber(state[2])

if tokens == nil then
	tokens = burst
end
if updated_at == nil then
	updated_at = now
end

local elapsed = math.max(0, now - updated_at) / 1000
tokens = math.min(burst, tokens + elapsed * rate)

local allowed = 0
if tokens >= 1 then
	tokens = tokens - 1
	allowed = 1
end

redis.call("HSET", key, "tokens", tokens, "updated_at", now)
redis.call("PEXPIRE", key, ttl)

return allowed
`

type Limiter struct {
	client *redis.Client
}

func NewLimiter(client *redis.Client) *Limiter {
	return &Limiter{client: client}
}

// AllowTokenBucket 令牌桶检查入口
func (l *Limiter) AllowTokenBucket(ctx context.Context, key string, ratePerSecond float64, burst int, ttl time.Duration) (bool, error) {
	if ratePerSecond <= 0 || burst <= 0 || ttl <= 0 {
		return false, errors.New("rate limit config invalid")
	}
	result, err := l.client.Eval(ctx, tokenBucketScript, []string{key}, ratePerSecond, burst, ttl.Milliseconds()).Int()
	if err != nil {
		return false, errors.WithStack(err)
	}
	return result == 1, nil
}
