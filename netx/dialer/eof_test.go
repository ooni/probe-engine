package dialer

import (
	"context"
	"io"
	"net"
)

type EOFDialer struct{}

func (EOFDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return nil, io.EOF
}

type EOFConnDialer struct{}

func (EOFConnDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return EOFConn{}, nil
}

type EOFConn struct {
	net.Conn
}

func (EOFConn) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (EOFConn) Write(p []byte) (int, error) {
	return 0, io.EOF
}

func (EOFConn) Close() error {
	return io.EOF
}

func (EOFConn) LocalAddr() net.Addr {
	return EOFAddr{}
}

type EOFAddr struct{}

func (EOFAddr) Network() string {
	return "tcp"
}

func (EOFAddr) String() string {
	return "127.0.0.1:1234"
}
