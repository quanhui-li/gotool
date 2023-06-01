package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	MaxDeleteNum = 1000
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
	// Close 关闭缓存
	Close() error
}

type BuildInMapCache struct {
	// map缓存数据
	data map[string]*Value
	// Cache读多写少，用读写锁
	mutex sync.RWMutex
	// Close 接收关闭信号
	close chan struct{}
	// 控制只关闭一次
	sync.Once
}

func NewBuildInMapCache(interval time.Duration) *BuildInMapCache {
	res := &BuildInMapCache{
		data:  make(map[string]*Value),
		mutex: sync.RWMutex{},
		close: make(chan struct{}),
	}

	// 开启goroutine定期循环删除过期的key，map遍历是随机的，不是顺序遍历，不存在每次轮训的key一样的情况，更优雅
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				counter := 0
				res.mutex.Lock()
				for key, val := range res.data {
					if counter > MaxDeleteNum {
						break
					}
					if val.deadlineBefore(t) {
						delete(res.data, key)
					}

					counter++
				}
				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()

	return res
}

// Set 设置key、value和过期时间，到了过期时间直接清除数据，如果不传过期时间则永不过期
func (b *BuildInMapCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.data[key] = &Value{
		value:    value,
		deadline: time.Now().Add(expiration),
	}

	//// 通过time.After的方式清理过期的数据，但是新建的key都会有阻塞的goroutine去清理，大量的goroutine被浪费了，不够优雅
	//if expiration > 0 {
	//	time.AfterFunc(expiration, func() {
	//		b.mutex.Lock()
	//		defer b.mutex.Unlock()
	//		val, ok := b.data[key]
	//		if ok && !val.deadline.IsZero() && val.deadline.Before(time.Now()) {
	//			delete(b.data, key)
	//		}
	//	})
	//}

	return nil
}

// Get 根据key查询出对应的数据，key设置有过期时间，goroutine有轮询时间，所以在下一次轮询前过期的key，查询的时候
// 会查询到，需要在Get方法中过过期处理
func (b *BuildInMapCache) Get(ctx context.Context, key string) (value any, err error) {
	b.mutex.RLock()
	val, ok := b.data[key]
	b.mutex.RUnlock()
	if !ok {
		return nil, NotExisted
	}

	// 考虑并发的情况，当一个goroutine Get拿到锁获取到key、val后，key是过期的，需要删除，另一个goroutine
	// 此时并发写了这个key，就不应该再删除了，而是应该返回新的val，这里需要二次检查key的过期时间(double check)
	now := time.Now()
	if !val.deadlineBefore(now) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		val, ok = b.data[key]
		if !ok {
			return nil, NotExisted
		}
		if !val.deadlineBefore(now) {
			delete(b.data, key)
		}
	}

	return val.value, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.data, key)

	return nil
}

func (b *BuildInMapCache) Close() error {
	b.Once.Do(func() {
		b.close <- struct{}{}
	})
	return nil
}

type Value struct {
	value    any
	deadline time.Time
}

func (v *Value) deadlineBefore(t time.Time) bool {
	return !v.deadline.IsZero() && v.deadline.Before(t)
}
