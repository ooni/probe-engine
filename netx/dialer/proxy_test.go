package dialer_test

import (
	"context"
	"errors"
	"io"
	"net/url"
	"testing"

	"github.com/ooni/probe-engine/netx/dialer"
)

func TestUnitProxyDialerDialContextNoProxyURL(t *testing.T) {
	expected := errors.New("mocked error")
	d := dialer.ProxyDialer{
		Dialer: dialer.FakeDialer{Err: expected},
	}
	conn, err := d.DialContext(context.Background(), "tcp", "www.google.com:443")
	if !errors.Is(err, expected) {
		t.Fatal(err)
	}
	if conn != nil {
		t.Fatal("conn is not nil")
	}
}

func TestUnitProxyDialerDialContextInvalidScheme(t *testing.T) {
	d := dialer.ProxyDialer{
		Dialer:   dialer.FakeDialer{},
		ProxyURL: &url.URL{Scheme: "antani"},
	}
	conn, err := d.DialContext(context.Background(), "tcp", "www.google.com:443")
	if err.Error() != "Scheme is not socks5" {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("conn is not nil")
	}
}

func TestUnitProxyDialerDialContextWithEOF(t *testing.T) {
	d := dialer.ProxyDialer{
		Dialer: dialer.FakeDialer{
			Err: io.EOF,
		},
		ProxyURL: &url.URL{Scheme: "socks5"},
	}
	conn, err := d.DialContext(context.Background(), "tcp", "www.google.com:443")
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("conn is not nil")
	}
}

func TestUnitProxyDialerDialContextWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately fail
	d := dialer.ProxyDialer{
		Dialer: dialer.FakeDialer{
			Err: io.EOF,
		},
		ProxyURL: &url.URL{Scheme: "socks5"},
	}
	conn, err := d.DialContext(ctx, "tcp", "www.google.com:443")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("conn is not nil")
	}
}

func TestUnitProxyDialerDialContextWithDialerSuccess(t *testing.T) {
	d := dialer.ProxyDialer{
		Dialer: dialer.FakeDialer{
			Conn: &dialer.FakeConn{
				ReadError:  io.EOF,
				WriteError: io.EOF,
			},
		},
		ProxyURL: &url.URL{Scheme: "socks5"},
	}
	conn, err := d.DialContextWithDialer(
		context.Background(), dialer.ProxyDialerWrapper{
			Dialer: d.Dialer,
		}, "tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestUnitProxyDialerDialContextWithDialerCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Stop immediately. The FakeDialer sleeps for some microseconds so
	// it is much more likely we immediately exit with done context. The
	// arm where we receive the conn is much less likely.
	cancel()
	d := dialer.ProxyDialer{
		Dialer: dialer.FakeDialer{
			Conn: &dialer.FakeConn{
				ReadError:  io.EOF,
				WriteError: io.EOF,
			},
		},
		ProxyURL: &url.URL{Scheme: "socks5"},
	}
	conn, err := d.DialContextWithDialer(
		ctx, dialer.ProxyDialerWrapper{
			Dialer: d.Dialer,
		}, "tcp", "www.google.com:443")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

func TestUnitProxyDialerWrapper(t *testing.T) {
	d := dialer.ProxyDialerWrapper{
		Dialer: dialer.FakeDialer{
			Err: io.EOF,
		},
	}
	conn, err := d.Dial("tcp", "www.google.com:443")
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("conn is not nil")
	}
}
