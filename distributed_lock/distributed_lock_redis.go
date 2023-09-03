package distributed_lock

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrFailedToRaceLock = errors.New("抢锁失败")
	ErrLockNotHold      = errors.New("你没有持有锁")
)

//go:embed lua/unlock.lua
var unlockScript string

//go:embed lua/refresh_lock.lua
var refreshScript string

// RedisDistributedLock 基于Redis实现的分布式锁
type RedisDistributedLock struct {
	client redis.Cmdable
}

func NewRedisDistributedLock(client redis.Cmdable) *RedisDistributedLock {
	return &RedisDistributedLock{
		client: client,
	}
}

// TryLock 尝试抢锁，key是存储在Redis中的键，
func (l *RedisDistributedLock) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := l.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrFailedToRaceLock
	}

	return &Lock{
		key:        key,
		val:        val,
		client:     l.client,
		expiration: expiration,
	}, nil
}

// Lock 锁
type Lock struct {
	// 存储在redis中的key
	key string
	// 锁的唯一标识，防止释放掉别人的锁
	val string
	// 过期时间，用于手动续约刷新过期时间
	expiration time.Duration
	// redis
	client redis.Cmdable
}

// Refresh 手动给锁续约
func (l Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, refreshScript, []string{l.key}, l.val, l.expiration.Seconds()).Int64()
	if err != nil {
		return err
	}

	if res != 1 {
		return ErrLockNotHold
	}

	return nil
}

// Unlock 解锁，因为TryLock返回的是*Lock，所以直接定义为Lock的方法
// 解锁过程涉及到并发问题，可以利用Redis单线程的特性使用脚本来完成解锁流程
func (l Lock) Unlock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, unlockScript, []string{l.key}, l.val).Int64()
	if err != nil {
		return err
	}

	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}
