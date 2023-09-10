package task_pool

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	p := NewPool(10, 100)
	defer func(p *Pool) {
		_ = p.Close()
	}(p)

	for i := 0; i <= 100; i++ {
		if err := p.Submit(context.Background(), func() {
			fmt.Println("测试程序")
		}); err != nil {
			t.Log(err)
		}
	}
	time.Sleep(10 * time.Millisecond)
}

func TestSyncPool(t *testing.T) {
	p := &sync.Pool{
		New: func() any {
			t.Log("创建资源了")
			// 最好永远不要返回nil
			return "hello"
		},
	}

	res := p.Get()
	val, _ := res.(string)
	t.Log(val)
	p.Put(val)

	res = p.Get()
	val, _ = res.(string)
	t.Log(val)

	res = p.Get()
	val, _ = res.(string)
	t.Log(val)

}
