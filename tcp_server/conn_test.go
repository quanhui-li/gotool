package conn

import (
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestConn(t *testing.T) {
	addr := ":8082"
	network := "tcp"
	go func() {
		if err := NewServer(network, addr).Start(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	client, err := NewClient(network, addr, 3*time.Second)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 1000; i++ {
		if err = client.Send("hello"); err != nil {
			panic(err)
		}

		if errors.Is(err, net.ErrClosed) || err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			panic(err)
		}

		if err != nil {
			fmt.Println(err)
			continue
		}

		msg, er := client.Receive()
		if er != nil {
			panic(err)
		}

		if errors.Is(er, net.ErrClosed) || er == io.EOF || errors.Is(er, io.ErrUnexpectedEOF) {
			panic(err)
		}

		if er != nil {
			fmt.Println(er)
			continue
		}
		fmt.Println(string(msg))
	}
}
