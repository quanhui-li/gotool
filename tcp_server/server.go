package conn

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	numberOfMessageLength = 8
)

type Server struct {
	network string
	addr    string
}

func NewServer(network, addr string) *Server {
	return &Server{
		network: network,
		addr:    addr,
	}
}

func (s Server) Start() error {
	listener, err := net.Listen(s.network, s.addr)
	if err != nil {
		return err
	}
	for {
		conn, er := listener.Accept()
		if er != nil {
			return err
		}

		go func() {
			if er = s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s Server) handleConn(conn net.Conn) error {
	errCh := make(chan error)
	defer func() {
		close(errCh)
	}()
	for {
		// 读数据
		bs := make([]byte, numberOfMessageLength)
		_, err := conn.Read(bs)
		if errors.Is(err, net.ErrClosed) || err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			return err
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		length := binary.BigEndian.Uint64(bs)
		msgBs := make([]byte, length)
		_, err = conn.Read(msgBs)
		if errors.Is(err, net.ErrClosed) || err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			return err
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 处理和发送数据
		go func() {
			data := s.handleData(msgBs)
			newBs := make([]byte, numberOfMessageLength+len(data))
			binary.BigEndian.PutUint64(newBs[:numberOfMessageLength], uint64(len(data)))
			copy(newBs[numberOfMessageLength:], data)
			_, err = conn.Write(newBs)
			if errors.Is(err, net.ErrClosed) || err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				errCh <- err
			}
			if err != nil {
				fmt.Println(err)
			}
		}()
		select {
		case err = <-errCh:
			if errors.Is(err, net.ErrClosed) || err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				return err
			}
		default:
			// 默认进入下一轮
		}

	}
}

func (s Server) handleData(data []byte) []byte {
	bs := make([]byte, 2*len(data))
	copy(bs, data)
	copy(bs[len(data):], data)
	return bs
}
