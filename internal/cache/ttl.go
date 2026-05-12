package cache

import (
	"math/rand"
	"time"
)

const (
	UserInfoTTL     = 30 * time.Minute
	FileListTTL     = 5 * time.Minute
	FileMetaTTL     = 15 * time.Minute
	FileIdentityTTL = time.Hour
	FileSampleTTL   = time.Hour
	ShareTokenTTL   = 30 * time.Minute
)

const ttlJitterRatio = 10

// JitterTTL 随机抖动TTL
func JitterTTL(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	maxJitter := base / ttlJitterRatio
	if maxJitter <= 0 {
		return base
	}
	return base + time.Duration(rand.Int63n(int64(maxJitter)+1))
}

// JitterTTLWithin 带上限的随机抖动TTL
func JitterTTLWithin(base time.Duration, max time.Duration) time.Duration {
	if max <= 0 {
		return max
	}
	ttl := JitterTTL(base)
	if ttl > max {
		return max
	}
	return ttl
}
