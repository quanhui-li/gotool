package channel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBroker_Send(t *testing.T) {
	broker := NewBroker()
	topicList := []string{"first_topic", "second_topic"}

	// 往不同的topic中各发送2000条消息
	for _, topic := range topicList {
		tp := topic
		go func() {
			for i := 0; i < 2000; i++ {
				if err := broker.Send(Msg{
					Content: time.Now().String(),
					Topic:   tp,
				}); err != nil {
					t.Log(err)
					return
				}

				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	// 创建6个goroutine订阅两个主题
	var wg sync.WaitGroup
	wg.Add(6)
	for i := 1; i <= 3; i++ {
		brokerName := fmt.Sprintf("first topic broker %d", i)
		go func() {
			defer wg.Done()
			msgs, err := broker.Subscribe("first_topic", 100)
			if err != nil {
				t.Log(err)
				return
			}

			for {
				select {
				case msg, ok := <-msgs:
					if !ok {
						return
					}
					data, _ := msg.(string)
					fmt.Printf("%s 消费到消息%s\n", brokerName, data)
				case <-time.After(time.Second):
					//t.Log("超时")
					return
				}
			}
		}()
	}

	for i := 1; i <= 3; i++ {
		brokerName := fmt.Sprintf("second topic broker %d", i)
		go func() {
			defer wg.Done()
			msgs, err := broker.Subscribe("second_topic", 100)
			if err != nil {
				t.Log(err)
				return
			}

			for {
				select {
				case msg, ok := <-msgs:
					if !ok {
						return
					}
					data, _ := msg.(string)
					fmt.Printf("%s 消费到消息%s\n", brokerName, data)
				case <-time.After(time.Second):
					//t.Log("超时")
					return
				}
			}
		}()
	}

	wg.Wait()
}
