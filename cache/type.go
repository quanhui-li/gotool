package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key不存在")
)

type Cache interface {
	// Set 设置缓存数据
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	// Get 获取缓存数据
	Get(ctx context.Context, key string) (any, error)
	// Delete 删除指定的缓存数据
	Delete(ctx context.Context, key string) error

	// LoadAndDelete 加载并删除数据
	LoadAndDelete(ctx context.Context, key string) (any, error)
}
