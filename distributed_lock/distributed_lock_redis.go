package distributed_lock

import (
	"context"
	_ "embed"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrFailedToLockRace = errors.New("抢锁失败")
	ErrLockNotExist     = errors.New("锁不存在，无需解锁")
	ErrFailedToUnlock   = errors.New("解锁失败")
	//go:embed lua/unlock.lua
	unlockVar string
)

// RedisDistributedLock 基于Redis实现的分布式锁
type RedisDistributedLock struct {
	client redis.Cmdable
}

// TryLock 尝试加锁，这里需要注意锁必须有过期时间，不能无限加锁，同时为保证锁的唯一性，使用uuid来标识
func (l *RedisDistributedLock) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := l.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrFailedToLockRace
	}

	return &Lock{
		key:    key,
		val:    val,
		client: l.client,
	}, nil
}

type Lock struct {
	// 存储的key
	key string
	// 存储的值，标识锁的唯一性
	val string
	// redis.Cmdable用于解锁
	client redis.Cmdable
}

// Unlock 解锁方法作为锁结构体的方法，使用lua脚本实现，防止并发环境下解锁的不一致问题
func (l *Lock) Unlock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, unlockVar, []string{l.key}, l.val).Int()
	if err != nil {
		return err
	}

	if res == 0 {
		// 锁根本不存在
		return ErrLockNotExist
	}

	if res != 1 {
		return ErrFailedToUnlock
	}

	return nil
}
