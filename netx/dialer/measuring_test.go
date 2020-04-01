package dialer

import (
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
)

func TestIntegrationMeasuringConn(t *testing.T) {
	conn := net.Conn(&MeasuringConn{
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
