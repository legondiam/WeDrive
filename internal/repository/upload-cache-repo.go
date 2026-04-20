package repository

import (
	"context"
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
}

func NewUploadCacheRepo(client *redis.Client) *UploadCacheRepo {
	return &UploadCacheRepo{client: client}
}

// partHashesKey 返回分块哈希有序集合 key。
func (r *UploadCacheRepo) partHashesKey(sessionID uint) string {
	return fmt.Sprintf("upload_session:%d:part_hashes", sessionID)
}

// partEtagsKey 返回分块ETag有序集合 key。
func (r *UploadCacheRepo) partEtagsKey(sessionID uint) string {
	return fmt.Sprintf("upload_session:%d:part_etags", sessionID)
}

// encodePartState 将分块编号和值编码为 Redis member，避免仅用值时无法区分重复分块哈希/ETag。
func encodePartState(partNumber int, value string) string {
	return fmt.Sprintf("%d|%s", partNumber, value)
}

// decodePartState 将 Redis member 解析回分块编号和值。
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

// setSortedPartValue 按 partNumber 写入有序集合，并刷新该 key 的过期时间。
func (r *UploadCacheRepo) setSortedPartValue(ctx context.Context, key string, partNumber int, value string, expire time.Duration) error {
	//先删除已经存在的分块
	number := strconv.FormatInt(int64(partNumber), 10)
	pipe := r.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, number, number)

	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(partNumber),
		Member: encodePartState(partNumber, value),
	})
	//设置过期时间
	pipe.Expire(ctx, key, expire)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// getSortedPartValues 读取指定有序集合中的全部分块状态，并按 score 顺序还原为结构化结果。
func (r *UploadCacheRepo) getSortedPartValues(ctx context.Context, key string) ([]UploadPartState, error) {
	values, err := r.client.ZRange(ctx, key, 0, -1).Result()
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
	values, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: number,
		Max: number,
	}).Result()
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

// SetPartHash 保存某个分块的哈希值
func (r *UploadCacheRepo) SetPartHash(ctx context.Context, sessionID uint, partNumber int, hash string, expire time.Duration) error {
	return r.setSortedPartValue(ctx, r.partHashesKey(sessionID), partNumber, hash, expire)
}

// SetPartETag 保存某个分块的 ETag
func (r *UploadCacheRepo) SetPartETag(ctx context.Context, sessionID uint, partNumber int, etag string, expire time.Duration) error {
	return r.setSortedPartValue(ctx, r.partEtagsKey(sessionID), partNumber, etag, expire)
}

// ListPartHashes 按序列出当前上传会话的全部分块哈希
func (r *UploadCacheRepo) ListPartHashes(ctx context.Context, sessionID uint) ([]UploadPartState, error) {
	return r.getSortedPartValues(ctx, r.partHashesKey(sessionID))
}

// GetPartHash 读取某个分块的哈希值。
func (r *UploadCacheRepo) GetPartHash(ctx context.Context, sessionID uint, partNumber int) (string, bool, error) {
	return r.getPartValue(ctx, r.partHashesKey(sessionID), partNumber)
}

// ListPartETags 按序列出当前上传会话的全部分块 ETag
func (r *UploadCacheRepo) ListPartETags(ctx context.Context, sessionID uint) ([]UploadPartState, error) {
	return r.getSortedPartValues(ctx, r.partEtagsKey(sessionID))
}

// GetPartETag 读取某个分块的 ETag。
func (r *UploadCacheRepo) GetPartETag(ctx context.Context, sessionID uint, partNumber int) (string, bool, error) {
	return r.getPartValue(ctx, r.partEtagsKey(sessionID), partNumber)
}

// ListUploadedParts 返回成功回报 ETag 的分块列表，用于断点续传时跳过已上传分块。
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

// DeleteUploadState 删除某个上传会话在 Redis 中保存的全部临时分块状态。
func (r *UploadCacheRepo) DeleteUploadState(ctx context.Context, sessionID uint) error {
	err := r.client.Del(ctx, r.partHashesKey(sessionID), r.partEtagsKey(sessionID)).Err()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
