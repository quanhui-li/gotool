package cache

import (
	"context"
	"time"
)

type Cache interface {
	// Set 设置缓存数据
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	// Get 获取缓存数据
	Get(ctx context.Context, key string) (any, error)
	// Delete 删除指定的缓存数据
	Delete(ctx context.Context, key string) error
	// Close 关闭缓存
	Close() error
}
