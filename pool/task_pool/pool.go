package task_pool

import "context"

type TaskFunc func()

type Pool struct {
	taskQueue chan TaskFunc
	close     chan struct{}
}

func NewPool(numG, capacity int) *Pool {
	res := &Pool{
		taskQueue: make(chan TaskFunc, capacity),
		close:     make(chan struct{}),
	}

	for i := 0; i < numG; i++ {
		go func() {
			for {
				select {
				case <-res.close:
					return
				case t := <-res.taskQueue:
					t()
				}
			}
		}()
	}

	return res
}

// Submit 往任务队列中推送任务
func (p *Pool) Submit(ctx context.Context, task TaskFunc) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.taskQueue <- task:
	}

	return nil
}

// Close 关闭任务队列
func (p *Pool) Close() error {
	close(p.close)
	return nil
}
