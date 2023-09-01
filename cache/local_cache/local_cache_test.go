package local_cache

import (
	"context"
	"github.com/go-playground/assert/v2"
	"strconv"
	"testing"
	"time"
)

func TestBindInMapCache_Set_Get_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		val     any
		cache   func(int) *BuildInMapCache
		wantErr error
		wantVal any
	}{
		{
			name: "key not exist",
			key:  "key not exist",
			cache: func(capacity int) *BuildInMapCache {
				cache := NewBuildInMapCache(capacity)
				return cache
			},
			wantErr: ErrKeyNotFound,
		},
		{
			name: "key exist",
			key:  "key exist",
			cache: func(capacity int) *BuildInMapCache {
				cache := NewBuildInMapCache(capacity)
				_ = cache.Set(context.Background(), "key exist", 100, 10*time.Second)
				return cache
			},
			wantErr: nil,
			wantVal: 100,
		},
		{
			name: "delete key",
			key:  "delete key",
			cache: func(capacity int) *BuildInMapCache {
				cache := NewBuildInMapCache(capacity)
				_ = cache.Set(context.Background(), "delete key", 100, 10*time.Second)
				val, err := cache.Get(context.Background(), "delete key")
				assert.Equal(t, err, nil)
				assert.Equal(t, val, 100)
				_ = cache.Delete(context.Background(), "delete key")
				return cache
			},
			wantErr: ErrKeyNotFound,
			wantVal: 100,
		},
		{
			name: "delete not exist key",
			key:  "delete not exist key",
			cache: func(capacity int) *BuildInMapCache {
				cache := NewBuildInMapCache(capacity)
				err := cache.Delete(context.Background(), "delete not exist key")
				assert.Equal(t, ErrKeyNotFound, err)
				return cache
			},
			wantErr: ErrKeyNotFound,
			wantVal: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache(100)
			val, err := cache.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				t.Log(err)
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func TestBuildInMapCache_Close(t *testing.T) {
	cache := NewBuildInMapCache(100)
	_ = cache.Set(context.Background(), "1", "1", time.Second)
	if err := cache.Close(); err != nil {
		t.Log(err)
	}

	if err := cache.Close(); err != nil {
		t.Log(err)
	}

	if err := cache.Close(); err != nil {
		t.Log(err)
	}
}

func TestBuildInMapCache_LoopDelete(t *testing.T) {
	wantCnt := 1
	cnt := 0
	cache := NewBuildInMapCache(100, BuildInMapCacheWithOnEvicted(func(key string, val any) {
		cnt++
	}))
	_ = cache.Set(context.Background(), "1", "1", 5*time.Second)
	val, err := cache.Get(context.Background(), "1")
	assert.Equal(t, nil, err)
	assert.Equal(t, "1", val)
	time.Sleep(11 * time.Second)
	assert.Equal(t, cnt, wantCnt)
}

func TestBuildInMapCache(t *testing.T) {
	cache := NewBuildInMapCache(10000)
	for i := 0; i < 10000; i++ {
		if err := cache.Set(context.Background(), strconv.Itoa(i), i, 100*time.Second); err != nil {
			t.Log(err)
		}
		t.Log("设置完成")
	}

	for i := 10000; i > 0; i-- {
		time.Sleep(10 * time.Millisecond)
		val, err := cache.Get(context.Background(), strconv.Itoa(i))
		if err != nil {
			t.Log(err)
			continue
		} else {
			t.Log("val: ", val)
		}
	}
	if err := cache.Close(); err != nil {
		t.Log(err)
	}

	if err := cache.Close(); err != nil {
		t.Log(err)
	}
}
