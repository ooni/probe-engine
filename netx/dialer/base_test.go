package dialer

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
)

func TestIntegrationBaseDialerSuccess(t *testing.T) {
	dialer := newBaseDialer()
	conn, err := dialer.DialContext(context.Background(), "tcp", "8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationBaseDialerErrorNoConnect(t *testing.T) {
	dialer := newBaseDialer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // stop immediately
	conn, err := dialer.DialContext(ctx, "tcp", "8.8.8.8:53")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if ctx.Err() == nil {
		t.Fatal("expected context to be expired here")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

// see whether we implement the interface
func newBaseDialer() Dialer {
	return EmitterDialer{
		Dialer: ErrWrapperDialer{
			Dialer: TimeoutDialer{
				Dialer: new(net.Dialer),
			},
		},
	}
}

func TestIntegrationEmitterConn(t *testing.T) {
	conn := net.Conn(&EmitterConn{
		Conn:    fakeconn{},
		Handler: handlers.NoHandler,
	})
	defer conn.Close()
	data := make([]byte, 1<<17)
	n, err := conn.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatal("invalid number of bytes read")
	}
	n, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatal("invalid number of bytes written")
	}
}

type fakeconn struct{}

func (fakeconn) Read(b []byte) (n int, err error) {
	n = len(b)
	return
}
func (fakeconn) Write(b []byte) (n int, err error) {
	n = len(b)
	return
}
func (fakeconn) Close() (err error) {
	return
}
func (fakeconn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}
func (fakeconn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}
func (fakeconn) SetDeadline(t time.Time) (err error) {
	return
}
func (fakeconn) SetReadDeadline(t time.Time) (err error) {
	return
}
func (fakeconn) SetWriteDeadline(t time.Time) (err error) {
	return
}
