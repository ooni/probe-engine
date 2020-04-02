package resolver

import (
	"context"
	"net"
	"time"
)

type FakeDialer struct {
	Conn net.Conn
	Err  error
}

func (d FakeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.Conn, d.Err
}

type FakeConn struct {
	ReadError             error
	ReadCount             int
	SetDeadlineError      error
	SetReadDeadlineError  error
	SetWriteDeadlineError error
	WriteError            error
}

func (c FakeConn) Read(b []byte) (n int, err error) {
	if c.ReadError != nil {
		return 0, c.ReadError
	}
	n = c.ReadCount
	return
}

func (c FakeConn) Write(b []byte) (n int, err error) {
	if c.WriteError != nil {
		return 0, c.WriteError
	}
	n = len(b)
	return
}

func (FakeConn) Close() (err error) {
	return
}

func (FakeConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (FakeConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c FakeConn) SetDeadline(t time.Time) (err error) {
	return c.SetDeadlineError
}

func (c FakeConn) SetReadDeadline(t time.Time) (err error) {
	return c.SetReadDeadlineError
}

func (c FakeConn) SetWriteDeadline(t time.Time) (err error) {
	return c.SetWriteDeadlineError
}
