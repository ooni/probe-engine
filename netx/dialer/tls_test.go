package dialer_test

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/dialer"
)

func TestIntegrationTLSDialerSuccess(t *testing.T) {
	dialer := dialer.NewTLSDialer(new(net.Dialer), new(tls.Config))
	conn, err := dialer.DialTLSContext(context.Background(), "tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("connection is nil")
	}
	conn.Close()
}

func TestIntegrationTLSDialerNilConfig(t *testing.T) {
	dialer := dialer.NewTLSDialer(new(net.Dialer), nil)
	conn, err := dialer.DialTLSContext(context.Background(), "tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("connection is nil")
	}
	conn.Close()
}

func TestIntegrationTLSDialerFailureSplitHostPort(t *testing.T) {
	dialer := dialer.NewTLSDialer(new(net.Dialer), new(tls.Config))
	conn, err := dialer.DialTLSContext(context.Background(), "tcp", "www.google.com") // missing port
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
	dialer := dialer.NewTLSDialer(new(net.Dialer), new(tls.Config))
	conn, err := dialer.DialTLSContext(ctx, "tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationTLSDialerFailureTLSHandshakeTimeout(t *testing.T) {
	dialer := &dialer.TLSDialer{
		Config: new(tls.Config),
		Dialer: new(net.Dialer),
		TLSHandshaker: dialer.TLSHandshakerSystem{
			HandshakeTimeout: time.Microsecond,
		},
	}
	conn, err := dialer.DialTLSContext(context.Background(), "tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestIntegrationTLSDialerFailureTLSHandshakeFailure(t *testing.T) {
	dialer := dialer.NewTLSDialer(new(net.Dialer), new(tls.Config))
	conn, err := dialer.DialTLSContext(context.Background(), "tcp", "self-signed.badssl.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}
