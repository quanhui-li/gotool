package queue

import (
	"errors"
	"sync"
	"time"
)

const (
	MaxSize         = 1024 * 1024 * 1024
	DefaultCapacity = 1000
)

var (
	TopicErr     = errors.New("topic not existed")
	QueueFullErr = errors.New("queue full")
	MaxSizeErr   = errors.New("exceeded max size")
)

type Broker struct {
	// 加锁保证map安全
	sync.RWMutex
	// 队列链，key是topic，value是队列
	queueLinks map[string][]chan []byte
	// 发送超时时长限制
	timeout time.Duration
	// 单条消息最大长度限制
	maxSize int
}

func NewBroker(timeout time.Duration) *Broker {
	return &Broker{
		queueLinks: make(map[string][]chan []byte),
		timeout:    timeout,
		maxSize:    MaxSize,
	}
}

// Send 发送消息到topic下的队列中
func (b *Broker) Send(msg Msg) error {
	if len(msg.Content) > b.maxSize {
		return MaxSizeErr
	}
	b.RLock()
	defer b.RUnlock()
	queueLink, ok := b.queueLinks[msg.Topic]
	if !ok {
		return TopicErr
	}

	for _, queue := range queueLink {
		select {
		case queue <- msg.Content:
		case <-time.After(b.timeout):
			return QueueFullErr
		}
	}

	return nil
}

// Subscribe 订阅主题和队列
func (b *Broker) Subscribe(topic string, capacity int) (<-chan []byte, error) {
	b.Lock()
	defer b.Unlock()

	if capacity <= 0 {
		capacity = DefaultCapacity
	}

	newQueue := make(chan []byte, capacity)
	_, ok := b.queueLinks[topic]
	if !ok {
		b.queueLinks[topic] = []chan []byte{newQueue}
	} else {
		b.queueLinks[topic] = append(b.queueLinks[topic], newQueue)
	}

	return newQueue, nil
}

// Close 关闭订阅的topic下的所有队列
func (b *Broker) Close(topic string) error {
	b.Lock()
	queue := b.queueLinks[topic]
	b.queueLinks[topic] = nil
	b.Unlock()

	// TODO 解决不传递topic直接关闭对应topic下的队列问题
	// 避免重复关闭的问题
	for _, queue := range queue {
		close(queue)
	}

	return nil
}

type Msg struct {
	// 发送消息的内容
	Content []byte
	// 订阅的队列主题
	Topic string
}
