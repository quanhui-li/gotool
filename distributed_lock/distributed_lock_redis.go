package distributed_lock

import (
	"context"
	_ "embed"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrFailedToRaceLock = errors.New("抢锁失败")
	ErrLockNotHold      = errors.New("你没有持有锁")
	ErrOverMaxCount     = errors.New("超过重试次数")
)

//go:embed lua/unlock.lua
var unlockScript string

//go:embed lua/refresh_lock.lua
var refreshScript string

//go:embed lua/lock.lua
var lockScript string

// RedisDistributedLock 基于Redis实现的分布式锁
type RedisDistributedLock struct {
	client redis.Cmdable
}

func NewRedisDistributedLock(client redis.Cmdable) *RedisDistributedLock {
	return &RedisDistributedLock{
		client: client,
	}
}

func (l *RedisDistributedLock) Lock(ctx context.Context, key string,
	timeout, expiration time.Duration,
	strategy RetryStrategy) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	for {
		c, cancel := context.WithTimeout(ctx, timeout)
		res, err := l.client.Eval(c, lockScript, []string{key}, []any{val, expiration.Seconds()}).Result()
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		interval, ok := strategy.Next(err)
		if !ok {
			return nil, ErrOverMaxCount
		}

		if res == "OK" {
			return &Lock{
				key:        key,
				val:        val,
				client:     l.client,
				expiration: expiration,
			}, nil
		}

		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}

		select {
		case <-timer.C:
			// 进行下一轮重试
		case <-ctx.Done():
			// context超时
			return nil, ctx.Err()
		}
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
	// unlockCh 通知停止自动续约的channel
	unlockCh chan struct{}
	// timeoutCh context超时后通知重试的channel
	timeoutCh chan struct{}
	// once 防止多次释放锁
	once *sync.Once
}

// AutoRefresh 自动续约机制，timeout是每次调用redis的context超时时间，interval是每次续约的间隔时间
func (l *Lock) AutoRefresh(interval, timeout time.Duration) error {
	l.timeoutCh = make(chan struct{}, 1)
	l.unlockCh = make(chan struct{}, 1)
	l.once = &sync.Once{}

	defer close(l.timeoutCh)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			// 这里是正常的续约逻辑处理
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					// 超时的业务员逻辑
					l.timeoutCh <- struct{}{}
					continue
				} else {
					return err
				}
			}
		case <-l.timeoutCh:
			// 这里是context超时后处理重试的逻辑
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					// 超时的业务员逻辑
					l.timeoutCh <- struct{}{}
					continue
				} else {
					return err
				}
			}
		case <-l.unlockCh:
			return nil
		}
	}
}

// Refresh 手动给锁续约
func (l *Lock) Refresh(ctx context.Context) error {
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
func (l *Lock) Unlock(ctx context.Context) error {
	var unlockErr error
	l.once.Do(func() {
		defer func() {
			select {
			case l.unlockCh <- struct{}{}:
				close(l.unlockCh)
			default:
				// 没有人调用自动续约，不需要处理
			}
		}()

		res, err := l.client.Eval(ctx, unlockScript, []string{l.key}, l.val).Int64()
		if err != nil {
			unlockErr = err
		}

		if res != 1 {
			unlockErr = ErrLockNotHold
		}
		return
	})

	return unlockErr
}
