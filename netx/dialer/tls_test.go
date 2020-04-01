package dialer

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/modelx"
)

func TestIntegrationTLSDialerSuccess(t *testing.T) {
	dialer := newTLSDialer()
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("connection is nil")
	}
	conn.Close()
}

func TestIntegrationTLSDialerSuccessWithMeasuringConn(t *testing.T) {
	dialer := newTLSDialer()
	dialer.(*TLSDialer).dialer = new(net.Dialer)
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("connection is nil")
	}
	conn.Close()
}

func TestIntegrationTLSDialerFailureSplitHostPort(t *testing.T) {
	dialer := newTLSDialer()
	conn, err := dialer.DialTLS("tcp", "www.google.com") // missing port
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationTLSDialerFailureConnectTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cause immediate timeout
	dialer := newTLSDialer()
	conn, err := dialer.DialTLSContext(ctx, "tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationTLSDialerFailureTLSHandshakeTimeout(t *testing.T) {
	dialer := newTLSDialer()
	dialer.(*TLSDialer).TLSHandshakeTimeout = 10 * time.Microsecond
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationTLSDialerFailureSetDeadline(t *testing.T) {
	dialer := newTLSDialer()
	dialer.(*TLSDialer).setDeadline = func(conn net.Conn, t time.Time) error {
		return errors.New("mocked error")
	}
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func newTLSDialer() modelx.TLSDialer {
	return NewTLSDialer(new(net.Dialer), new(tls.Config))
}
