package cache

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"time"
)

type loadFuncType func(ctx context.Context, key string) (any, error)

// ReadThroughCache 需要用户必须赋值loadFunc、expiration、logFunc，如果不赋值，会发生Panic
type ReadThroughCache struct {
	Cache
	// 需要用户传入，用于从数据库加载数据
	loadFunc loadFuncType
	// 过期时间设置
	expiration time.Duration
	// 记录日志的方法
	logFunc func(msg string)
	// 引入singleFlight，合并多条相同的请求，减轻数据库压力
	g *singleflight.Group
}

func NewReadThroughCache(logFunc func(string), loadFunc loadFuncType, expiration time.Duration) *ReadThroughCache {
	return &ReadThroughCache{
		loadFunc:   loadFunc,
		logFunc:    logFunc,
		expiration: expiration,
	}
}

// SyncGet 缓存对外听过的获取数据方法
func (r *ReadThroughCache) SyncGet(ctx context.Context, key string) (any, error) {
	res, err := r.Cache.Get(ctx, key)
	if err == nil {
		return res, nil
	}

	// 未找到数据
	if errors.Is(err, ErrKeyNotFound) {
		val, er, _ := r.g.Do(key, func() (interface{}, error) {
			val, er := r.loadFunc(ctx, key)
			if er != nil {
				return val, er
			}

			er = r.Cache.Set(ctx, key, val, r.expiration)
			// 缓存刷新失败不应该影响到数据返回的逻辑，可以让调用放传入一个写日志方法，
			// 记录缓存数据失败的原因，也可以直接返回错误信息，不过推荐记录日志
			if er != nil {
				r.logFunc(fmt.Sprintf("写入缓存数据失败，错误为: %s", er))
			}
			return val, nil
		})
		return val, er
	}

	return nil, err
}

// SemiAsyncGet 半异步获取缓存数据
func (r *ReadThroughCache) SemiAsyncGet(ctx context.Context, key string) (any, error) {
	res, err := r.Cache.Get(ctx, key)
	if err == nil {
		return res, nil
	}

	// 缓存中不存在，查询数据库
	val, er := r.loadFunc(ctx, key)
	if er != nil {
		return val, er
	}

	// 开启goroutine刷新缓存，goroutine中是无法返回错误的，所以必须使用日志记录
	go func() {
		er = r.Cache.Set(ctx, key, val, r.expiration)
		if er != nil {
			r.logFunc(fmt.Sprintf("写入缓存数据失败，错误为: %s", er.Error()))
		}
	}()

	return val, nil
}

// AsyncGet 完全异步的加载数据、刷新缓存
func (r *ReadThroughCache) AsyncGet(ctx context.Context, key string) (any, error) {
	res, err := r.Cache.Get(ctx, key)
	if err == nil {
		return res, nil
	}

	// 缓存没有数据需要异步加载
	go func() {
		val, er := r.loadFunc(ctx, key)
		if er != nil {
			r.logFunc(fmt.Sprintf("从数据库读取数据失败，错误为: %s", err.Error()))
			return
		}

		er = r.Cache.Set(ctx, key, val, r.expiration)
		if er != nil {
			r.logFunc(fmt.Sprintf("刷新缓存数据失败，错误为: %s", err.Error()))
			return
		}
	}()

	return res, err
}
