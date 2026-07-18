package repository

import (
	"WeDrive/internal/cache"
	"WeDrive/internal/cacheguard"
	"WeDrive/internal/config"
	"context"
	"encoding/binary"
	"hash/fnv"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type BloomRepo struct {
	client *redis.Client
	guard  *cacheguard.RedisGuard
	bits   uint64
	hashes uint64
}

func NewBloomRepo(client *redis.Client, guard *cacheguard.RedisGuard) *BloomRepo {
	return NewBloomRepoWithConfig(client, guard, uint64(config.GlobalConf.Bloom.Bits), uint64(config.GlobalConf.Bloom.Hashes))
}

// NewBloomRepoWithConfig 创建布隆过滤器仓储。
func NewBloomRepoWithConfig(client *redis.Client, guard *cacheguard.RedisGuard, bits uint64, hashes uint64) *BloomRepo {
	if bits == 0 {
		bits = 95850584
	}
	if hashes == 0 {
		hashes = 7
	}
	return &BloomRepo{client: client, guard: guard, bits: bits, hashes: hashes}
}

// Add 写入一个布隆过滤器元素。
func (r *BloomRepo) Add(ctx context.Context, name string, item string) error {
	offsets := r.offsets(item)
	return r.guard.Do(ctx, func(ctx context.Context) error {
		pipe := r.client.Pipeline()
		key := cache.BloomKey(name)
		for _, offset := range offsets {
			pipe.SetBit(ctx, key, int64(offset), 1)
		}
		_, err := pipe.Exec(ctx)
		return errors.WithStack(err)
	})
}

// AddMany 批量写入布隆过滤器元素。
func (r *BloomRepo) AddMany(ctx context.Context, name string, items []string) error {
	if len(items) == 0 {
		return nil
	}
	return r.guard.Do(ctx, func(ctx context.Context) error {
		pipe := r.client.Pipeline()
		key := cache.BloomKey(name)
		for _, item := range items {
			for _, offset := range r.offsets(item) {
				pipe.SetBit(ctx, key, int64(offset), 1)
			}
		}
		_, err := pipe.Exec(ctx)
		return errors.WithStack(err)
	})
}

// MightContain 判断元素是否可能存在。
func (r *BloomRepo) MightContain(ctx context.Context, name string, item string) (bool, error) {
	offsets := r.offsets(item)
	values, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) ([]*redis.IntCmd, error) {
		pipe := r.client.Pipeline()
		key := cache.BloomKey(name)
		cmds := make([]*redis.IntCmd, 0, len(offsets))
		for _, offset := range offsets {
			cmds = append(cmds, pipe.GetBit(ctx, key, int64(offset)))
		}
		_, err := pipe.Exec(ctx)
		return cmds, err
	})
	if err != nil {
		return true, errors.WithStack(err)
	}
	for _, cmd := range values {
		bit, err := cmd.Result()
		if err != nil {
			return true, errors.WithStack(err)
		}
		if bit == 0 {
			return false, nil
		}
	}
	return true, nil
}

// SetReady 设置布隆过滤器是否可用于拦截请求。
func (r *BloomRepo) SetReady(ctx context.Context, name string, ready bool) error {
	value := "0"
	if ready {
		value = "1"
	}
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return errors.WithStack(r.client.Set(ctx, cache.BloomReadyKey(name), value, 0).Err())
	})
}

// IsReady 判断布隆过滤器是否可用。
func (r *BloomRepo) IsReady(ctx context.Context, name string) (bool, error) {
	value, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (string, error) {
		return r.client.Get(ctx, cache.BloomReadyKey(name)).Result()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return value == "1", nil
}

// Clear 删除布隆过滤器位图和 ready 标记。
func (r *BloomRepo) Clear(ctx context.Context, name string) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return errors.WithStack(r.client.Del(ctx, cache.BloomKey(name), cache.BloomReadyKey(name)).Err())
	})
}

// Touch 设置布隆过滤器 key 的保活时间。
func (r *BloomRepo) Touch(ctx context.Context, name string, ttl time.Duration) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		return errors.WithStack(r.client.Expire(ctx, cache.BloomKey(name), ttl).Err())
	})
}

// offsets 计算元素对应的位图偏移量。
func (r *BloomRepo) offsets(item string) []uint64 {
	h1 := hashWithSeed(item, 0)
	h2 := hashWithSeed(item, 1)
	if h2 == 0 {
		h2 = 1
	}
	offsets := make([]uint64, 0, r.hashes)
	for i := uint64(0); i < r.hashes; i++ {
		offsets = append(offsets, (h1+i*h2)%r.bits)
	}
	return offsets
}

func hashWithSeed(item string, seed uint64) uint64 {
	h := fnv.New64a()
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], seed)
	_, _ = h.Write(buf[:])
	_, _ = h.Write([]byte(item))
	return h.Sum64()
}
