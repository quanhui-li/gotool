package cache

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestBuildInMapCache(t *testing.T) {
	cache := NewBuildInMapCache(1000000)
	for i := 0; i < 800000; i++ {
		if err := cache.Set(context.Background(), strconv.Itoa(i), i, 600*time.Second); err != nil {
			t.Log(err)
		}
		t.Log("设置完成")
	}

	for i := 0; i < 800000; i++ {
		time.Sleep(10 * time.Millisecond)
		val, err := cache.Get(context.Background(), strconv.Itoa(i))
		if err != nil {
			t.Log(err)
			continue
		}
		t.Log("val: ", val)
	}
	if err := cache.Close(); err != nil {
		t.Log(err)
	}

	if err := cache.Close(); err != nil {
		t.Log(err)
	}
}
