package cacheguard

import (
	"WeDrive/internal/config"
	"WeDrive/pkg/logger"
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sony/gobreaker/v2"
)

const (
	defaultInterval            = 10 * time.Second
	defaultBucketPeriod        = time.Second
	defaultMinRequests         = 20
	defaultFailureRate         = 0.5
	defaultConsecutiveFailures = 5
	defaultTimeout             = 5 * time.Second
	defaultHalfOpenRequests    = 2
)

type RedisGuard struct {
	enabled bool
	breaker *gobreaker.CircuitBreaker[any]
}

// NewRedisGuard 创建 Redis 熔断保护器。
func NewRedisGuard() *RedisGuard {
	conf := normalizeConfig(config.GlobalConf.RedisBreaker)
	if !conf.Enabled {
		return &RedisGuard{}
	}
	settings := gobreaker.Settings{
		Name:         "redis",
		MaxRequests:  uint32(conf.HalfOpenMaxRequests),
		Interval:     conf.Interval,
		BucketPeriod: conf.BucketPeriod,
		Timeout:      conf.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.ConsecutiveFailures >= uint32(conf.ConsecutiveFailures) {
				return true
			}
			if counts.Requests < uint32(conf.MinRequests) {
				return false
			}
			return float64(counts.TotalFailures)/float64(counts.Requests) >= conf.FailureRate
		},
		IsSuccessful: func(err error) bool {
			return !IsRedisFailure(err)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			if logger.S != nil {
				logger.S.Warnf("%s 熔断状态变化: %s -> %s", name, from.String(), to.String())
			}
		},
	}
	return &RedisGuard{
		enabled: true,
		breaker: gobreaker.NewCircuitBreaker[any](settings),
	}
}

// Do 在 Redis 熔断保护下执行一次操作。
func (g *RedisGuard) Do(ctx context.Context, fn func(context.Context) error) error {
	_, err := DoResult(g, ctx, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}

// DoResult 在 Redis 熔断保护下执行一次带返回值的操作。
func DoResult[T any](g *RedisGuard, ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	if g == nil || !g.enabled {
		return fn(ctx)
	}
	result, err := g.breaker.Execute(func() (any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return fn(ctx)
		}
	})
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return zero, ErrCircuitOpen
	}
	if err != nil {
		return zero, err
	}
	value, ok := result.(T)
	if !ok {
		return zero, errors.New("redis guard result type invalid")
	}
	return value, nil
}

// normalizeConfig 补齐 Redis 熔断配置默认值。
func normalizeConfig(conf config.RedisBreakerConf) config.RedisBreakerConf {
	if conf.Interval <= 0 {
		conf.Interval = defaultInterval
	}
	if conf.BucketPeriod <= 0 {
		conf.BucketPeriod = defaultBucketPeriod
	}
	if conf.MinRequests <= 0 {
		conf.MinRequests = defaultMinRequests
	}
	if conf.FailureRate <= 0 || conf.FailureRate > 1 {
		conf.FailureRate = defaultFailureRate
	}
	if conf.ConsecutiveFailures <= 0 {
		conf.ConsecutiveFailures = defaultConsecutiveFailures
	}
	if conf.Timeout <= 0 {
		conf.Timeout = defaultTimeout
	}
	if conf.HalfOpenMaxRequests <= 0 {
		conf.HalfOpenMaxRequests = defaultHalfOpenRequests
	}
	return conf
}
