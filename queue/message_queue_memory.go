package queue

import (
	"errors"
	"sync"
)

const (
	DefaultCapacity = 1000 // 默认容量
	MaxRetry        = 5    // 最大重试次数
)

type Broker struct {
	// 用于存放所有订阅消息的broker，为了针对topic做精细化的管控，使用了map结构
	// key是topic，val是切片，所有订阅了该消息的broker
	// 如果不关心topic，可以使用切片，每次消息过来发送到所有的broker上即可
	brokerChain map[string][]chan Message
	// 加锁保护
	mu sync.RWMutex
	// 重试策略，默认不重试，只发送一次，如果消息队列已满，该消息队列将跳过本次消息发送，
	// 如果开启重启，最多会重试5次
	Retry bool
}

// Send 发送消息
func (b *Broker) Send(msg Message) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.brokerChain[msg.Topic] {
		if b.Retry {
			for i := 0; i < MaxRetry; i++ {
				select {
				case ch <- msg:
					break
				default:
				}
			}
			return errors.New("消息队列已满")
		} else {
			select {
			case ch <- msg:
			default:
				return errors.New("消息队列已满")
			}
		}
	}

	return nil
}

// Subscribe 订阅一个消息管道，用于接受消息
// @param topic 是订阅消息的主题
// @param capacity 接受消息的管道的容量，容量是可选的，如果传有容量，则使用用户传的，如果没有传则使用默认的
// @return <-chan Message 一个只读的channel
// @return error 错误
func (b *Broker) Subscribe(topic string, capacity ...int) (<-chan Message, error) {
	var cap int
	if len(capacity) == 0 {
		cap = DefaultCapacity
	} else {
		cap = capacity[0]
	}
	ch := make(chan Message, cap)

	b.mu.Lock()
	defer b.mu.Unlock()
	chain, ok := b.brokerChain[topic]
	if !ok {
		b.brokerChain[topic] = []chan Message{ch}
	} else {
		chain = append(chain, ch)
	}

	return ch, nil
}

// Close 关闭订阅消息的队列
func (b *Broker) Close(topic string) {
	b.mu.Lock()
	chain, ok := b.brokerChain[topic]
	if !ok {
		return
	}
	b.brokerChain[topic] = nil // 避免了重复关闭问题
	b.mu.Unlock()

	for _, ch := range chain {
		close(ch)
	}
}

type Message struct {
	// 订阅的主题
	Topic string
	// 消息的内容
	Content any
}
