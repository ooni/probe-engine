package errwrapping_test

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/errwrapping"
	"github.com/ooni/probe-engine/netx/logging"
	"github.com/ooni/probe-engine/netx/measurable"
)

func TestIntegrationLookupHostFailure(t *testing.T) {
	ops := errwrapping.ErrWrapper{Operations: measurable.Defaults{}}
	ctx := measurable.WithOperations(context.Background(), ops)
	req, err := http.NewRequestWithContext(ctx, "GET", "http://xx.facebook.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := measurable.DefaultHTTPClient.Do(req)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resp != nil {
		t.Fatal("expected nil error here")
	}
	t.Logf("%s", errwrapping.Failure(err))
	t.Logf("%s", errwrapping.Operation(err))
}

func TestIntegrationLookupTLSHandshakeFailure(t *testing.T) {
	ops := errwrapping.ErrWrapper{Operations: tlsHandshakerWithServerName{
		Operations: measurable.Defaults{},
		ServerName: "www.google.com",
	}}
	ctx := measurable.WithOperations(context.Background(), ops)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://facebook.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := measurable.DefaultHTTPClient.Do(req)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resp != nil {
		t.Fatal("expected nil error here")
	}
	t.Logf("%s", errwrapping.Failure(err))
	t.Logf("%s", errwrapping.Operation(err))
}

type tlsHandshakerWithServerName struct {
	measurable.Operations
	ServerName string
}

func (th tlsHandshakerWithServerName) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	config.ServerName = th.ServerName
	return th.Operations.Handshake(ctx, conn, config)
}

func TestIntegrationLookupConnectFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	log.SetLevel(log.DebugLevel)
	ctx = measurable.WithOperations(ctx, logging.Handler{
		Operations: errwrapping.ErrWrapper{
			Operations: connectorWithErr{
				Operations: measurable.Defaults{},
			},
		},
		Logger: log.Log,
	})
	req, err := http.NewRequestWithContext(ctx, "GET", "http://facebook.com:111", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := measurable.DefaultHTTPClient.Do(req)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resp != nil {
		t.Fatal("expected nil error here")
	}
	t.Logf("%s", errwrapping.Failure(err))
	t.Logf("%s", errwrapping.Operation(err))
}

type connectorWithErr struct {
	measurable.Operations
}

func (connectorWithErr) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	return nil, io.EOF
}
