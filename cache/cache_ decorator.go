package cache

import (
	"errors"
	"sync/atomic"
	"time"
)

var (
	OverCapacityErr = errors.New("cache: 超过容量限制")
)

type LocalCache struct {
	*BuildInCache
	// maxCount 最多可以存储多少个key
	maxCount int32
	// curCount 当前已经存储key的数量
	curCount int32
}

func NewLocalCache(c *BuildInCache, mnt int32) *LocalCache {
	lc := &LocalCache{
		BuildInCache: c,
		maxCount:     mnt,
	}

	// 通过回调函数来减少key的数量，删除的时候回调用回调方法，用于处理减少数量
	// 需要在传入的回调函数上包上一层
	origin := c.onEvicted
	lc.onEvicted = func(key string, value any) {
		atomic.AddInt32(&lc.curCount, -1)
		if c.onEvicted != nil {
			origin(key, value)
		}
	}

	return lc
}

func (l *LocalCache) Set(key string, value any, operation time.Duration) error {
	//// 这种方式如果key已经存在，那么计数就是错误的
	//cnt := atomic.AddUint32(&l.curCount, 1)
	//if cnt > l.maxCount {
	//	atomic.AddUint32(&l.curCount, -1)
	//	return OverCapacityErr
	//}
	//return l.Set(key, value, operation)

	// 这种方式在写数据的时候会出现并发问题，计数的时候加锁保护了，但是计数后又释放掉锁，在写数据的时候可能
	// 会有其他的goroutine来计数
	//l.Lock()
	//_, ok := l.data[key]
	//if !ok {
	//	atomic.AddUint32(&l.curCount, 1)
	//}
	//if atomic.LoadUint32(&l.curCount) > l.maxCount {
	//	l.Unlock()
	//	return OverCapacityErr
	//}
	//l.Unlock()
	//return l.Set(key, value, operation)

	//
	l.Lock()
	defer l.Unlock()
	_, ok := l.data[key]
	if !ok {
		atomic.AddInt32(&l.curCount, 1)
	}
	if atomic.LoadInt32(&l.curCount) > l.maxCount {
		atomic.AddInt32(&l.maxCount, -1)
		return OverCapacityErr
	}
	return l.set(key, value, operation)
}
