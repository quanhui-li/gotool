package cache

import (
	"context"
	"errors"
	_ "github.com/golang/mock/mockgen/model"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrFailToSetKey = errors.New("cache: 设置key失败")
)

type RedisCache struct {
	client redis.Cmdable
}

func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	res, err := r.client.Set(ctx, key, value, expiration).Result()
	if err != nil {
		return err
	}
	if res != "OK" {
		return ErrFailToSetKey
	}

	return nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (any, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}
