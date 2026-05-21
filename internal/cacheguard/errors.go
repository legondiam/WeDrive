package cacheguard

import (
	"context"
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

var ErrCircuitOpen = errors.New("redis circuit breaker open")

// IsCircuitOpen 判断错误是否为 Redis 熔断打开。
func IsCircuitOpen(err error) bool {
	return errors.Is(err, ErrCircuitOpen)
}

// IsRedisFailure 判断错误是否应计入 Redis 熔断失败。
func IsRedisFailure(err error) bool {
	if err == nil || errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "pool timeout") ||
		strings.Contains(msg, "readonly") ||
		strings.Contains(msg, "loading") ||
		strings.Contains(msg, "clusterdown")
}
