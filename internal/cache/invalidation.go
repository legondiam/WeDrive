package cache

import (
	"WeDrive/pkg/logger"
	"context"
	"time"
)

const DelayedDeleteDelay = time.Second

// DelayedDelete 延迟删除缓存
func DelayedDelete(delay time.Duration, deleteFunc func(context.Context) error) {
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()

		<-timer.C

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := deleteFunc(ctx); err != nil {
			logger.S.Warnf("延迟删除缓存失败:%v", err)
		}
	}()
}
