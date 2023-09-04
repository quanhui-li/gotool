package conn

import (
	"encoding/binary"
	"net"
	"time"
)

type Client struct {
	conn net.Conn
}

func NewClient(network, addr string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn: conn,
	}, nil
}

func (c Client) Send(msg string) error {
	bs := make([]byte, numberOfMessageLength+len(msg))
	binary.BigEndian.PutUint64(bs[:numberOfMessageLength], uint64(len(msg)))
	copy(bs[numberOfMessageLength:], msg)
	_, err := c.conn.Write(bs)
	if err != nil {
		_ = c.conn.Close()
	}

	return nil
}

func (c Client) Receive() (string, error) {
	bs := make([]byte, numberOfMessageLength)
	_, err := c.conn.Read(bs)
	if err != nil {
		return "", err
	}

	length := binary.BigEndian.Uint64(bs)
	data := make([]byte, length)
	_, err = c.conn.Read(data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
