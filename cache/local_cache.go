package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	NotExisted = errors.New("key not existed")
)

type MapCacheInter interface {
	// Set 设置缓存数据
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	// Get 查询单条数据
	Get(ctx context.Context, key string) (value any, err error)
	// Delete 删除指定key的数据
	Delete(ctx context.Context, key string) error
}

type BuildInMapCache struct {
	// map缓存数据
	data map[string]*Value
	// Cache读多写少，用读写锁
	mutex sync.RWMutex
}

func NewBuildInMapCache() *BuildInMapCache {
	return &BuildInMapCache{
		data:  make(map[string]*Value),
		mutex: sync.RWMutex{},
	}
}

// Set 设置key、value和过期时间，到了过期时间直接清除数据，如果不传过期时间则永不过期
func (b *BuildInMapCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.data[key] = &Value{
		value:    value,
		deadline: time.Now().Add(expiration),
	}

	// 通过time.After的方式清理过期的数据，但是新建的key都会有阻塞的goroutine去清理，大量的goroutine被浪费了，不够优雅
	// TODO 优化过期设置
	if expiration > 0 {
		time.AfterFunc(expiration, func() {
			b.mutex.Lock()
			defer b.mutex.Unlock()
			val, ok := b.data[key]
			if ok && !val.deadline.IsZero() && val.deadline.Before(time.Now()) {
				delete(b.data, key)
			}
		})
	}

	return nil
}

func (b *BuildInMapCache) Get(ctx context.Context, key string) (value any, err error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	val, ok := b.data[key]
	if !ok {
		return nil, NotExisted
	}

	return val, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.data, key)

	return nil
}

type Value struct {
	value    any
	deadline time.Time
}
