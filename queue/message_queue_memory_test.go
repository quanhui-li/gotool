package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	FirstCustomer  = "消费者 1"
	SecondCustomer = "消费者 2"
	ThirdCustomer  = "消费者 3"
)

func TestQueue(t *testing.T) {
	qu := NewBroker()
	topic := "1234"
	err := qu.Subscribe(topic, FirstCustomer)
	if err != nil {
		t.Log("订阅错误", err)
		return
	}

	err = qu.Subscribe(topic, SecondCustomer)
	if err != nil {
		t.Log("订阅错误", err)
		return
	}

	err = qu.Subscribe(topic, ThirdCustomer)
	if err != nil {
		t.Log("订阅错误", err)
		return
	}
	var wg sync.WaitGroup

	go func() {
		defer func() {
			qu.Close(topic)
		}()
		for i := 0; i < 10000; i++ {
			errQueue, ok := qu.ErrQueue(topic)
			if !ok {
				t.Log("错误消息队列不存在")
				return
			}
			err := qu.Send(Message{
				Topic:   topic,
				Content: time.Now().String(),
			})
			if err != nil {
				t.Log("topic不存在")
				return
			}

			select {
			case msg, ok := <-errQueue:
				if !ok {
					t.Log("topic已取消订阅")
					return
				}
				fmt.Println("错误消息: ", msg)
				if msg.Err.Error() == "消息队列已满" {
					fmt.Println("开始重试")
					if err = qu.Send(msg.Message); err != nil {
						t.Log(err)
					}
				}
				time.Sleep(10 * time.Millisecond)
			default:
				// 空的 case 分支
			}
		}
	}()

	// 先订阅topic，三个goroutine
	wg.Add(3)
	for i := 0; i < 3; i++ {
		name := ""
		switch i {
		case 0:
			name = FirstCustomer
		case 1:
			name = SecondCustomer
		case 2:
			name = ThirdCustomer
		}

		go func(name string) {
			defer wg.Done()

			count := 0
			defer func() {
				t.Logf("%s接收到%d条消息", name, count)
			}()

			ch, ok := qu.Queue(topic, name)
			if !ok {
				t.Log("队列不存在")
				return
			}

			for {
				//if count >= 100 && name == ThirdCustomer {
				//	qu.ClosesQueue(topic, ThirdCustomer)
				//}
				select {
				case msg, ok := <-ch:
					if !ok {
						t.Log("通道已关闭")
						return
					}

					t.Logf(fmt.Sprintf("%s接收到消息%s", name, msg.Content))
					count++
				}
			}
		}(name)
	}

	// 发送消息
	wg.Wait()
}
