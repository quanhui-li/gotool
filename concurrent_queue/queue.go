package concurrent_queue

import (
	"context"
	"errors"
	"sync"
)

type ConcurrentQueue[T any] struct {
	queue []any
	mu    sync.Mutex
}

func NewConcurrentQueue[T any](size int) *ConcurrentQueue[T] {
	return &ConcurrentQueue[T]{
		queue: make([]any, size),
	}
}

func (c *ConcurrentQueue[T]) EnQueue(ctx context.Context, data T) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queue = append(c.queue, data)
	return nil
}

func (c *ConcurrentQueue[T]) DeQueue(ctx context.Context) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.IsEmpty() {
		var t T
		return t, errors.New("空的队列")
	}
	data := c.queue[0]
	c.queue = c.queue[1:]
	return data, nil
}

func (c *ConcurrentQueue[T]) IsFull() bool {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentQueue[T]) IsEmpty() bool {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentQueue[T]) Len() uint64 {
	//TODO implement me
	panic("implement me")
}
