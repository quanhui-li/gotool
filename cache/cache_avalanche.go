package cache

import (
	"context"
	"math/rand"
	"time"
)

type RandomExpirationCache struct {
	Cache
}

func (s *RandomExpirationCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	if expiration > 0 {
		offset := time.Duration(rand.Intn(100)) * time.Second
		return s.Cache.Set(ctx, key, val, expiration+offset)
	}
	return s.Cache.Set(ctx, key, val, expiration)
}
