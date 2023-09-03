//go:build e2e

package distributed_lock

import (
	"context"
	"github.com/go-playground/assert/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedisDistributedLock_TryLock_e2e(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "122.9.137.145:6319",
		Password: "123456",
	})
	testCases := []struct {
		name    string
		key     string
		before  func(t *testing.T)
		after   func(t *testing.T)
		wantErr error
		lock    *Lock
	}{
		{
			name: "lock be hold",
			key:  "key1",
			before: func(t *testing.T) {
				// 模拟别人有所
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := client.Set(ctx, "key1", "value1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := client.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, res, "value1")
			},
			wantErr: ErrFailedToRaceLock,
			lock: &Lock{
				key:    "key1",
				val:    "value1",
				client: client,
			},
		},
		{
			name:   "lock success",
			key:    "key2",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, err := client.GetDel(ctx, "key2").Result()
				require.NoError(t, err)
			},
			wantErr: nil,
			lock: &Lock{
				key:    "key2",
				client: client,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dl := NewRedisDistributedLock(client)
			tc.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			lock, err := dl.TryLock(ctx, tc.key, time.Minute)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.lock.key, lock.key)
			if lock.client == nil {
				return
			}
			tc.after(t)
		})
	}
}

func TestRedisDistributedLock_Unlock_e2e(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "122.9.137.145:6319",
		Password: "123456",
	})
	testCases := []struct {
		name    string
		key     string
		before  func(t *testing.T)
		after   func(t *testing.T)
		lock    *Lock
		wantErr error
	}{
		// 锁不存在
		{
			name:    "unlock failed",
			key:     "not exist",
			before:  func(t *testing.T) {},
			after:   func(t *testing.T) {},
			wantErr: ErrLockNotExist,
			lock: &Lock{
				key:    "not exist1",
				val:    "not exist1",
				client: client,
			},
		},
		// 锁存在，但不是自己的锁
		{
			name: "unlock hold",
			key:  "unlock hold",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				res, err := client.Set(ctx, "lock hold", "123456", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after:   func(t *testing.T) {},
			wantErr: ErrLockNotExist,
			lock: &Lock{
				key:    "lock hold",
				val:    "1232132131232132",
				client: client,
			},
		},
		// 锁存在且是自己的锁
		{
			name: "unlock success",
			key:  "unlock success",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				res, err := client.Set(ctx, "unlock success", "123456", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after:   func(t *testing.T) {},
			wantErr: nil,
			lock: &Lock{
				key:    "unlock success",
				val:    "123456",
				client: client,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := tc.lock.Unlock(ctx)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}
