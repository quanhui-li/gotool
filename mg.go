package channel

import (
	"errors"
	"sync"
)

var (
	QueueFullErr = errors.New("队列已满，消息发送失败")
	TopicErr     = errors.New("topic不存在")
)

// Broker Broker结构体
type Broker struct {
	// 读写锁
	sync.RWMutex
	// 队列链，map的key是订阅消息队列的主题Topic，Topic对应多个队列，每一个管道就是一个消息队列
	queueLinks map[string][]chan interface{}
}

func NewBroker() *Broker {
	return &Broker{
		queueLinks: make(map[string][]chan interface{}),
	}
}

// Send 给所有的消息队列发送消息
func (b *Broker) Send(msg Msg) error {
	b.RLock()
	defer b.RUnlock()
	queueLink, ok := b.queueLinks[msg.Topic]
	if !ok {
		return TopicErr
	}
	for _, queue := range queueLink {
		for {
			select {
			case queue <- msg:
				break
			default:
				return QueueFullErr
			}
		}
	}

	return nil
}

// Subscribe 订阅新的channel到队列链中，注册时需要指定队列的容量，
// 如果容量小于等于0，则默认赋予2000的容量。
func (b *Broker) Subscribe(topic string, capacity int) (<-chan interface{}, error) {
	b.Lock()
	defer b.Unlock()
	if capacity <= 0 {
		capacity = 2000
	}
	// 判断topic是否存在，不存在则创建新的topic，存在则在新的主题上添加一个队列
	newQueue := make(chan interface{}, capacity)
	_, ok := b.queueLinks[topic]
	if !ok {
		b.queueLinks[topic] = []chan interface{}{newQueue}
	} else {
		b.queueLinks[topic] = append(b.queueLinks[topic], newQueue)
	}

	return newQueue, nil
}

// Msg 消息结构体
type Msg struct {
	Content interface{}
	Topic   string
}
