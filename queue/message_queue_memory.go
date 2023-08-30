package queue

import (
	"errors"
	"sync"
)

const (
	DefaultCapacity = 10 // 默认容量
)

type Broker struct {
	// 用于存放所有订阅消息的broker，为了针对topic做精细化的管控，使用了map结构
	// key是topic，val是map，存储了所有订阅了该消息的broker，内部map的key是队列的名称，
	// val对应的是channel，订阅的队列
	// 如果不关心topic，可以使用切片，每次消息过来发送到所有的broker上即可
	brokerChain map[string]map[string]chan Message
	// 加锁保护
	mu sync.RWMutex
	// 返回发送错误消息的队列，每一个topic对应队列，这个队列是错误队列，用于返回发送失败的消息及错误的内容
	// 容量是默认容量
	topicErrQueue map[string]chan ErrMessage
}

func NewBroker() *Broker {
	return &Broker{
		brokerChain:   map[string]map[string]chan Message{},
		topicErrQueue: map[string]chan ErrMessage{},
	}
}

// Send 发送消息
func (b *Broker) Send(msg Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	chains, ok := b.brokerChain[msg.Topic]
	if !ok {
		return errors.New("topic不存在")
	}
	if msg.Queue != "" {
		_, ok := chains[msg.Queue]
		if !ok {
			b.topicErrQueue[msg.Topic] <- ErrMessage{
				Message: msg,
				Err:     errors.New("消息队列不存在"),
			}
		}
	}

	chains, ok = b.brokerChain[msg.Topic]
	if !ok {
		b.topicErrQueue[msg.Topic] <- ErrMessage{
			Message: msg,
			Err:     errors.New("消息队列不存在"),
		}
	}

	var wg sync.WaitGroup
	for ne, ch := range chains {
		// TODO 处理重试的数据没有写入到指定队列中的问题
		if msg.Queue != "" && ne != msg.Queue {
			continue
		}
		name := ne
		wg.Add(1)
		go func(ch chan Message, name string) {
			defer wg.Done()
			select {
			case ch <- msg:
				return
			default:
				msg.Queue = name
				b.topicErrQueue[msg.Topic] <- ErrMessage{
					Message: msg,
					Err:     errors.New("消息队列已满"),
				}
			}
		}(ch, name)
	}

	wg.Wait()
	return nil
}

// ErrQueue 获取topic下的错误消息队列
func (b *Broker) ErrQueue(topic string) (<-chan ErrMessage, bool) {
	queue, ok := b.topicErrQueue[topic]
	if !ok {
		return nil, false
	}
	return queue, true
}

// Queue 获取topic下的消息队列
func (b *Broker) Queue(topic, queue string) (<-chan Message, bool) {
	qu, ok := b.brokerChain[topic][queue]
	if !ok {
		return nil, false
	}
	return qu, true
}

// Subscribe 订阅一个消息管道，用于接受消息
// @param topic 是订阅消息的主题
// @param capacity 接受消息的管道的容量，容量是可选的，如果传有容量，则使用用户传的，如果没有传则使用默认的
// @return <-chan Message 一个只读的channel
// @return error 错误
func (b *Broker) Subscribe(topic, queueName string, capacity ...int) error {
	var cap int
	if len(capacity) == 0 {
		cap = DefaultCapacity
	} else {
		cap = capacity[0]
	}
	ch := make(chan Message, cap)
	errCh := make(chan ErrMessage, DefaultCapacity)

	b.mu.Lock()
	defer b.mu.Unlock()
	chain, ok := b.brokerChain[topic]
	if !ok {
		b.brokerChain[topic] = map[string]chan Message{
			queueName: ch,
		}
		b.topicErrQueue[topic] = errCh
	} else {
		_, ok := chain[queueName]
		if ok {
			return errors.New("队列已存在，请勿重复订阅")
		}
		chain[queueName] = ch
	}

	return nil
}

// Close 关闭订阅消息的队列
func (b *Broker) Close(topic string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	chain, ok := b.brokerChain[topic]
	if !ok {
		return
	}
	// 避免了重复关闭问题
	b.brokerChain[topic] = nil
	// 关闭错误消息的队列
	close(b.topicErrQueue[topic])

	for _, ch := range chain {
		close(ch)
	}
}

// ClosesQueue 关闭指定topic下的指定队列，先查询topic下是否存在该队列，存在则删除map，然后关闭channel
// 再查询删除之后的topic下是否还有队列，如果没有就删除错误topic，关闭错误队列
// @param topic 订阅的主题
// @param queue 队列的名称
func (b *Broker) ClosesQueue(topic, queue string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	chain, ok := b.brokerChain[topic]
	if !ok {
		return
	}
	// 先拿到channel
	ch, ok := chain[queue]
	if !ok {
		return
	}
	delete(chain, queue)
	close(ch)

	if len(b.brokerChain[topic]) == 0 {
		close(b.topicErrQueue[topic])
	}
}

type Message struct {
	// 订阅的主题
	Topic string
	// 订阅的队列
	Queue string
	// 消息的内容
	Content any
}

type ErrMessage struct {
	Message
	// 错误的信息
	Err error
}
