package tlsdialer

import (
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/internal/dialer/dialerbase"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestIntegrationSuccess(t *testing.T) {
	dialer := newdialer()
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("connection is nil")
	}
	conn.Close()
}

func TestIntegrationSuccessWithMeasuringConn(t *testing.T) {
	dialer := newdialer()
	dialer.(*TLSDialer).dialer = dialerbase.New(
		time.Now(), handlers.NoHandler, new(net.Dialer), 17,
	)
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("connection is nil")
	}
	conn.Close()
}

func TestIntegrationFailureSplitHostPort(t *testing.T) {
	dialer := newdialer()
	conn, err := dialer.DialTLS("tcp", "www.google.com") // missing port
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationFailureConnectTimeout(t *testing.T) {
	dialer := newdialer()
	dialer.(*TLSDialer).ConnectTimeout = 10 * time.Microsecond
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationFailureTLSHandshakeTimeout(t *testing.T) {
	dialer := newdialer()
	dialer.(*TLSDialer).TLSHandshakeTimeout = 10 * time.Microsecond
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationFailureSetDeadline(t *testing.T) {
	dialer := newdialer()
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

func newdialer() modelx.TLSDialer {
	return New(new(net.Dialer), new(tls.Config))
}
