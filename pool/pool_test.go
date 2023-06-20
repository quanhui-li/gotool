package pool

import (
	"context"
	"fmt"
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
