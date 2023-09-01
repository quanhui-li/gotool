//go:build e2e

package redis_cache

import (
	"context"
	"github.com/go-playground/assert/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedisCache_e2e_Set(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "122.9.137.145:6319",
		Password: "123456",
	})
	cache := NewRedisCache(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	val := "key1"
	err := cache.Set(ctx, "key1", val, time.Second)
	require.NoError(t, err)
	res, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, val, res)
}

func TestRedisCache_e2e_Set_Drive(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "122.9.137.145:6319",
		Password: "123456",
	})
	cache := NewRedisCache(client)

	testCases := []struct {
		name    string
		key     string
		wantErr error
		wantVal any
		before  func()
		after   func()
	}{
		{
			name:    "get key",
			key:     "get key",
			wantErr: nil,
			wantVal: "v1",
			before: func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := cache.Set(ctx, "get key", "v1", 10*time.Second)
				require.NoError(t, err)
			},
			after: func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := cache.Delete(ctx, "get key")
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			tc.before()
			val, err := cache.Get(ctx, "get key")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, val)
			tc.after()
		})
	}
}
