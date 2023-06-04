package cache

import (
	"errors"
	"sync"
	"time"
)

const (
	maxDeleteCount = 1000
)

var (
	ErrKeyNotExisted = errors.New("cache: key不存在")
)

type Options func(cache *BuildInCache)

func BuildInCacheWithOnEvicted(fn func(string, any)) Options {
	return func(cache *BuildInCache) {
		cache.onEvicted = fn
	}
}

type BuildInCache struct {
	// data map结构缓存数据
	data map[string]*Value
	// 读写锁控制并发
	sync.RWMutex
	// 关闭goroutine
	close chan struct{}
	// 防止重复关闭
	sync.Once
	// 注册回调方法处理上下文，可选
	onEvicted func(key string, value any)
}

func NewBuildInCache(capacity int, opts ...Options) *BuildInCache {
	res := &BuildInCache{
		data:      make(map[string]*Value, capacity),
		close:     make(chan struct{}),
		onEvicted: func(key string, value any) {},
	}

	for _, opt := range opts {
		opt(res)
	}

	// 开启goroutine定时轮询一部分key，map的轮询是随机的，不需要考虑每次轮询都是固定数据的问题
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case t := <-ticker.C:
				res.Lock()
				count := 0
				for key, val := range res.data {
					if count >= maxDeleteCount {
						break
					}
					if val.deadlineBefore(t) {
						res.delete(key)
					}
					count++
				}
				res.Unlock()
			case <-res.close:
				return
			}
		}
	}()

	return res
}

// Set 设置key、value和过期时间
func (b *BuildInCache) Set(key string, value any, expiration time.Duration) error {
	b.Lock()
	defer b.Unlock()
	return b.set(key, value, expiration)

	//// 通过time.AfterFunc设置过期
	//if expiration > 0 {
	//	time.AfterFunc(expiration, func() {
	//		b.Lock()
	//		defer b.Unlock()
	//		val, ok := b.data[key]
	//		now := time.Now()
	//		if ok && !val.deadline.IsZero() && val.deadline.Before(now) {
	//			delete(b.data, key)
	//		}
	//	})
	//}
}

func (b *BuildInCache) set(key string, value any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	b.data[key] = &Value{
		value:    value,
		deadline: dl,
	}
	return nil
}

// Get 根据key查询缓存的内容
func (b *BuildInCache) Get(key string) (any, error) {
	b.RLock()
	val, ok := b.data[key]
	b.RUnlock()
	if !ok {
		return nil, ErrKeyNotExisted
	}

	now := time.Now()
	// 存在就需要判断是否过期，过期就删除，需要二次检验，防止并发修改问题，解决轮询漏掉过期key的问题
	if val.deadlineBefore(now) {
		b.Lock()
		defer b.Unlock()
		val, ok = b.data[key]
		if !ok {
			return nil, ErrKeyNotExisted
		}
		if val.deadlineBefore(now) {
			b.delete(key)
		}
	}

	return val.value, nil
}

func (b *BuildInCache) Close() error {
	b.Once.Do(func() {
		b.close <- struct{}{}
	})
	return nil
}

// Delete 删除key
func (b *BuildInCache) Delete(key string) error {
	b.Lock()
	defer b.Unlock()
	b.delete(key)
	return nil
}

// delete 处理删除key和调用回调方法的程序
func (b *BuildInCache) delete(key string) {
	val, ok := b.data[key]
	if !ok {
		b.onEvicted(key, nil)
		return
	}
	delete(b.data, key)
	b.onEvicted(key, val.value)
}

type Value struct {
	// value消息内容
	value any
	// 过期时间
	deadline time.Time
}

func (v *Value) deadlineBefore(t time.Time) bool {
	return !v.deadline.IsZero() && v.deadline.Before(t)
}
