package concurrent_queue

import "context"

// Queue 队列对外提供的接口
type Queue[T any] interface {
	// EnQueue 消息入队
	EnQueue(ctx context.Context, data T) error
	// DeQueue 消息出队
	DeQueue(ctx context.Context) (T, error)
	// IsFull 队列是否已满
	IsFull() bool
	// IsEmpty 是否空队列
	IsEmpty() bool
	// Len 队列的长度
	Len() uint64
}
