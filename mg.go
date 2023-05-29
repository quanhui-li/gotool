package channel

import (
	"errors"
	"sync"
)

var (
	TopicErr     = errors.New("topic not existed")
	QueueFullErr = errors.New("queue full")
)

type Broker struct {
	// 加锁保证map安全
	sync.RWMutex
	// 队列链，key是topic，value是队列
	queueLinks map[string][]chan interface{}
}

func NewBroker() *Broker {
	return &Broker{
		queueLinks: make(map[string][]chan interface{}),
	}
}

// Send 发送消息到topic下的队列中
func (b *Broker) Send(msg Msg) error {
	b.RLock()
	defer b.RUnlock()
	queueLink, ok := b.queueLinks[msg.Topic]
	if !ok {
		return TopicErr
	}

	for _, queue := range queueLink {
		select {
		case queue <- msg.Content:
		default:
			return QueueFullErr
		}

	}

	return nil
}

func (b *Broker) Subscribe(topic string, capacity int) (<-chan interface{}, error) {
	b.Lock()
	defer b.Unlock()

	if capacity <= 0 {
		capacity = 2000
	}

	newQueue := make(chan interface{}, capacity)
	_, ok := b.queueLinks[topic]
	if !ok {
		b.queueLinks[topic] = []chan interface{}{newQueue}
	} else {
		b.queueLinks[topic] = append(b.queueLinks[topic], newQueue)
	}

	return newQueue, nil
}

type Msg struct {
	// 发送消息的内容
	Content interface{}
	// 订阅的队列主题
	Topic string
}
