package cacheguard

import (
	"WeDrive/internal/config"
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// TestRedisGuardOpenAfterFailures 验证连续 Redis 故障会打开熔断。
func TestRedisGuardOpenAfterFailures(t *testing.T) {
	config.GlobalConf.RedisBreaker = config.RedisBreakerConf{
		Enabled:             true,
		Interval:            time.Second,
		BucketPeriod:        100 * time.Millisecond,
		MinRequests:         100,
		FailureRate:         0.5,
		ConsecutiveFailures: 2,
		Timeout:             time.Second,
		HalfOpenMaxRequests: 1,
	}
	guard := NewRedisGuard()

	for i := 0; i < 2; i++ {
		err := guard.Do(context.Background(), func(ctx context.Context) error {
			return context.DeadlineExceeded
		})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("want deadline error, got %v", err)
		}
	}

	called := false
	err := guard.Do(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("want circuit open, got %v", err)
	}
	if called {
		t.Fatal("redis call should be skipped while circuit is open")
	}
}

// TestRedisNilDoesNotOpen 验证缓存未命中不会触发熔断。
func TestRedisNilDoesNotOpen(t *testing.T) {
	config.GlobalConf.RedisBreaker = config.RedisBreakerConf{
		Enabled:             true,
		Interval:            time.Second,
		BucketPeriod:        100 * time.Millisecond,
		MinRequests:         1,
		FailureRate:         0.1,
		ConsecutiveFailures: 1,
		Timeout:             time.Second,
		HalfOpenMaxRequests: 1,
	}
	guard := NewRedisGuard()

	for i := 0; i < 3; i++ {
		err := guard.Do(context.Background(), func(ctx context.Context) error {
			return redis.Nil
		})
		if !errors.Is(err, redis.Nil) {
			t.Fatalf("want redis nil, got %v", err)
		}
	}

	called := false
	err := guard.Do(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("want nil, got %v", err)
	}
	if !called {
		t.Fatal("redis call should not be skipped after redis.Nil")
	}
}
