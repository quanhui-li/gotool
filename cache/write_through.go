package cache

import (
	"context"
	"fmt"
	"time"
)

type writeFuncType func(ctx context.Context, key string, val any) error

// WriteThroughCache 写数据的缓存
type WriteThroughCache struct {
	Cache
	// writeFunc 数据写到数据库中的方法
	writeFunc writeFuncType
	// logFunc 记录日志的方法
	logFunc func(msg string)
}

func NewWriteThroughCache(writeFunc writeFuncType, logFunc func(string)) *WriteThroughCache {
	return &WriteThroughCache{
		writeFunc: writeFunc,
		logFunc:   logFunc,
	}
}

// SyncSet 同步的设置数据
func (w *WriteThroughCache) SyncSet(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 先更新数据库
	err := w.writeFunc(ctx, key, val)
	if err != nil {
		return err
	}

	// 更新缓存数据
	return w.Set(ctx, key, val, expiration)
}

// SemiSyncSet 半异步的设置数据
func (w *WriteThroughCache) SemiSyncSet(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 先更新数据库
	err := w.writeFunc(ctx, key, val)
	if err != nil {
		return err
	}

	// 更新缓存数据
	go func() {
		if err = w.Set(ctx, key, val, expiration); err != nil {
			w.logFunc(fmt.Sprintf("刷新缓存数据失败，错误为: %s", err.Error()))
		}
	}()
	return nil
}
