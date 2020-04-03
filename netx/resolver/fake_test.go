package resolver

import (
	"context"
	"io"
	"net"
	"testing"
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
	ReadData              []byte
	SetDeadlineError      error
	SetReadDeadlineError  error
	SetWriteDeadlineError error
	WriteError            error
}

func (c *FakeConn) Read(b []byte) (int, error) {
	if len(c.ReadData) > 0 {
		n := copy(b, c.ReadData)
		c.ReadData = c.ReadData[n:]
		return n, nil
	}
	if c.ReadError != nil {
		return 0, c.ReadError
	}
	return 0, io.EOF
}

func (c *FakeConn) Write(b []byte) (n int, err error) {
	if c.WriteError != nil {
		return 0, c.WriteError
	}
	n = len(b)
	return
}

func (*FakeConn) Close() (err error) {
	return
}

func (*FakeConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (*FakeConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c *FakeConn) SetDeadline(t time.Time) (err error) {
	return c.SetDeadlineError
}

func (c *FakeConn) SetReadDeadline(t time.Time) (err error) {
	return c.SetReadDeadlineError
}

func (c *FakeConn) SetWriteDeadline(t time.Time) (err error) {
	return c.SetWriteDeadlineError
}

type FakeTransport struct {
	Data []byte
	Err  error
}

func (ft FakeTransport) RoundTrip(ctx context.Context, query []byte) ([]byte, error) {
	return ft.Data, ft.Err
}

func (ft FakeTransport) RequiresPadding() bool {
	return false
}

func (ft FakeTransport) Address() string {
	return ""
}

func (ft FakeTransport) Network() string {
	return "fake"
}

type FakeEncoder struct {
	Data []byte
	Err  error
}

func (fe FakeEncoder) Encode(domain string, qtype uint16, padding bool) ([]byte, error) {
	return fe.Data, fe.Err
}

func TestUnitFakeResolverThatFails(t *testing.T) {
	client := NewFakeResolverThatFails()
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if addrs != nil {
		t.Fatal("expected nil here")
	}
}

func TestUnitFakeResolverWithResult(t *testing.T) {
	orig := []string{"10.0.0.1"}
	client := NewFakeResolverWithResult(orig)
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(orig) != len(addrs) || orig[0] != addrs[0] {
		t.Fatal("not the result we expected")
	}
}
