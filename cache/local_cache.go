package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key不存在")
)

var _ Cache = (*BuildInMapCache)(nil)

// BuildInMapCache 本地内存缓存
type BuildInMapCache struct {
	// 存储的数据，key是缓存的键，val是值，any类型
	data map[string]*value
	// 加锁保护资源
	mu sync.RWMutex
	// 关闭goroutine的channel
	close chan struct{}
	// 引入once防止重复关闭的问题
	once sync.Once
}

func NewBuildInMapCache(capacity int) *BuildInMapCache {
	cache := &BuildInMapCache{
		data:  make(map[string]*value, capacity),
		close: make(chan struct{}),
	}

	// 设置goroutine定时轮询过期的缓存数据
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for tk := range ticker.C {
			select {
			case <-cache.close:
				return
			default:
				cache.mu.Lock()
				count := 0
				for key, val := range cache.data {
					// 控制每次的轮询数量，防止轮询过多导致性能问题
					if count > 1000 {
						break
					}
					if val.timeout(tk) {
						delete(cache.data, key)
					}
					count++
				}
				cache.mu.Unlock()
			}
		}
	}()

	return cache
}

func (m *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}

	m.data[key] = &value{
		val:      val,
		deadline: dl,
	}

	return nil
}

func (m *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	m.mu.RLock()
	val, ok := m.data[key]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrKeyNotFound
	}

	t := time.Now()
	if val.timeout(t) {
		m.mu.Lock()
		defer m.mu.RUnlock()
		res, ok := m.data[key]
		if !ok {
			return nil, ErrKeyNotFound
		}
		if res.timeout(t) {
			delete(m.data, key)
			return nil, ErrKeyNotFound
		}
		return res, nil
	}
	return val, nil
}

func (m *BuildInMapCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *BuildInMapCache) Close() error {
	m.once.Do(func() {
		m.close <- struct{}{}
		close(m.close)
	})
	return nil
}

type value struct {
	// 存储的值
	val any
	// 过期时间
	deadline time.Time
}

func (v value) timeout(t time.Time) bool {
	return !v.deadline.IsZero() && v.deadline.Before(t)
}
