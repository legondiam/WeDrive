package repository

import (
	"WeDrive/internal/cacheguard"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type UploadPartState struct {
	PartNumber int
	Value      string
}

type UploadCacheRepo struct {
	client *redis.Client
	guard  *cacheguard.RedisGuard
}

const consumeInstantPrepareScript = `
local value = redis.call("GET", KEYS[1])
if value then
	redis.call("DEL", KEYS[1])
end
return value
`

func NewUploadCacheRepo(client *redis.Client, guard *cacheguard.RedisGuard) *UploadCacheRepo {
	return &UploadCacheRepo{client: client, guard: guard}
}

// partHashesKey 返回分块哈希有序集合 key。
func (r *UploadCacheRepo) partHashesKey(sessionID uint) string {
	return fmt.Sprintf("upload_session:%d:part_hashes", sessionID)
}

// partEtagsKey 返回分块 ETag 有序集合 key。
func (r *UploadCacheRepo) partEtagsKey(sessionID uint) string {
	return fmt.Sprintf("upload_session:%d:part_etags", sessionID)
}

// instantPrepareKey 返回秒传挑战准备状态 key。
func (r *UploadCacheRepo) instantPrepareKey(prepareID string) string {
	return fmt.Sprintf("instant_prepare:%s", prepareID)
}

// encodePartState 将分块编号和值编码成 Redis member。
func encodePartState(partNumber int, value string) string {
	return fmt.Sprintf("%d|%s", partNumber, value)
}

// decodePartState 将 Redis member 解码成分块状态。
func decodePartState(raw string) (UploadPartState, error) {
	part, value, ok := strings.Cut(raw, "|")
	if !ok {
		return UploadPartState{}, errors.New("分块状态格式错误")
	}
	partNumber, err := strconv.Atoi(part)
	if err != nil {
		return UploadPartState{}, errors.WithStack(err)
	}
	return UploadPartState{PartNumber: partNumber, Value: value}, nil
}

// setSortedPartValue 按分块编号写入有序集合并刷新过期时间。
func (r *UploadCacheRepo) setSortedPartValue(ctx context.Context, key string, partNumber int, value string, expire time.Duration) error {
	return r.guard.Do(ctx, func(ctx context.Context) error {
		number := strconv.FormatInt(int64(partNumber), 10)
		pipe := r.client.TxPipeline()
		pipe.ZRemRangeByScore(ctx, key, number, number)
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(partNumber),
			Member: encodePartState(partNumber, value),
		})
		pipe.Expire(ctx, key, expire)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
}

// getSortedPartValues 读取全部分块状态并按分块编号排序。
func (r *UploadCacheRepo) getSortedPartValues(ctx context.Context, key string) ([]UploadPartState, error) {
	values, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) ([]string, error) {
		return r.client.ZRange(ctx, key, 0, -1).Result()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []UploadPartState{}, nil
		}
		return nil, errors.WithStack(err)
	}
	parts := make([]UploadPartState, 0, len(values))
	for _, item := range values {
		partState, decodeErr := decodePartState(item)
		if decodeErr != nil {
			return nil, decodeErr
		}
		parts = append(parts, partState)
	}
	return parts, nil
}

// getPartValue 读取指定分块的状态值。
func (r *UploadCacheRepo) getPartValue(ctx context.Context, key string, partNumber int) (string, bool, error) {
	number := strconv.FormatInt(int64(partNumber), 10)
	values, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) ([]string, error) {
		return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: number,
			Max: number,
		}).Result()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, errors.WithStack(err)
	}
	if len(values) == 0 {
		return "", false, nil
	}
	partState, decodeErr := decodePartState(values[0])
	if decodeErr != nil {
		return "", false, decodeErr
	}
	return partState.Value, true, nil
}

// SetPartHash 保存分块哈希。
func (r *UploadCacheRepo) SetPartHash(ctx context.Context, sessionID uint, partNumber int, hash string, expire time.Duration) error {
	return r.setSortedPartValue(ctx, r.partHashesKey(sessionID), partNumber, hash, expire)
}

// SetPartETag 保存分块 ETag。
func (r *UploadCacheRepo) SetPartETag(ctx context.Context, sessionID uint, partNumber int, etag string, expire time.Duration) error {
	return r.setSortedPartValue(ctx, r.partEtagsKey(sessionID), partNumber, etag, expire)
}

// ListPartHashes 列出上传会话的全部分块哈希。
func (r *UploadCacheRepo) ListPartHashes(ctx context.Context, sessionID uint) ([]UploadPartState, error) {
	return r.getSortedPartValues(ctx, r.partHashesKey(sessionID))
}

// GetPartHash 读取指定分块哈希。
func (r *UploadCacheRepo) GetPartHash(ctx context.Context, sessionID uint, partNumber int) (string, bool, error) {
	return r.getPartValue(ctx, r.partHashesKey(sessionID), partNumber)
}

// ListPartETags 列出上传会话的全部分块 ETag。
func (r *UploadCacheRepo) ListPartETags(ctx context.Context, sessionID uint) ([]UploadPartState, error) {
	return r.getSortedPartValues(ctx, r.partEtagsKey(sessionID))
}

// GetPartETag 读取指定分块 ETag。
func (r *UploadCacheRepo) GetPartETag(ctx context.Context, sessionID uint, partNumber int) (string, bool, error) {
	return r.getPartValue(ctx, r.partEtagsKey(sessionID), partNumber)
}

// ListUploadedParts 返回已回报 ETag 的分块编号。
func (r *UploadCacheRepo) ListUploadedParts(ctx context.Context, sessionID uint) ([]int, error) {
	partStates, err := r.ListPartETags(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	partNumbers := make([]int, 0, len(partStates))
	for _, part := range partStates {
		partNumbers = append(partNumbers, part.PartNumber)
	}
	return partNumbers, nil
}

// DeleteUploadState 删除上传会话的 Redis 临时状态。
func (r *UploadCacheRepo) DeleteUploadState(ctx context.Context, sessionID uint) error {
	err := r.guard.Do(ctx, func(ctx context.Context) error {
		return r.client.Del(ctx, r.partHashesKey(sessionID), r.partEtagsKey(sessionID)).Err()
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SetInstantPrepare 保存秒传挑战准备状态。
func (r *UploadCacheRepo) SetInstantPrepare(ctx context.Context, prepareID string, payload any, expire time.Duration) error {
	return r.setJSONValue(ctx, r.instantPrepareKey(prepareID), payload, expire)
}

// GetInstantPrepare 读取秒传挑战准备状态。
func (r *UploadCacheRepo) GetInstantPrepare(ctx context.Context, prepareID string, target any) (bool, error) {
	return r.getJSONValue(ctx, r.instantPrepareKey(prepareID), target)
}

// ConsumeInstantPrepare 原子读取并删除秒传挑战准备状态。
func (r *UploadCacheRepo) ConsumeInstantPrepare(ctx context.Context, prepareID string, target any) (bool, error) {
	result, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) (any, error) {
		return r.client.Eval(ctx, consumeInstantPrepareScript, []string{r.instantPrepareKey(prepareID)}).Result()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	raw, ok := result.(string)
	if !ok {
		return false, errors.New("instant prepare state format invalid")
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

// DeleteInstantPrepare 删除秒传挑战准备状态。
func (r *UploadCacheRepo) DeleteInstantPrepare(ctx context.Context, prepareID string) error {
	return r.deleteKey(ctx, r.instantPrepareKey(prepareID))
}

// setJSONValue 写入 JSON 缓存值。
func (r *UploadCacheRepo) setJSONValue(ctx context.Context, key string, payload any, expire time.Duration) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := r.guard.Do(ctx, func(ctx context.Context) error {
		return r.client.Set(ctx, key, raw, expire).Err()
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// getJSONValue 读取 JSON 缓存值。
func (r *UploadCacheRepo) getJSONValue(ctx context.Context, key string, target any) (bool, error) {
	raw, err := cacheguard.DoResult(r.guard, ctx, func(ctx context.Context) ([]byte, error) {
		return r.client.Get(ctx, key).Bytes()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

// deleteKey 删除指定 Redis key。
func (r *UploadCacheRepo) deleteKey(ctx context.Context, key string) error {
	if err := r.guard.Do(ctx, func(ctx context.Context) error {
		return r.client.Del(ctx, key).Err()
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
