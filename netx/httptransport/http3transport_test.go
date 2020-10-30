package httptransport_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/selfcensor"
)

func TestUnitHTTP3TransportSuccess(t *testing.T) {
	txp := httptransport.NewHTTP3Transport(selfcensor.SystemDialer{}, dialer.TLSDialer{})

	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("unexpected nil response here")
	}
	if resp.StatusCode != 200 {
		t.Fatal("HTTP statuscode should be 200 OK", resp.StatusCode)
	}
}

func TestUnitHTTP3TransportFailure(t *testing.T) {
	txp := httptransport.NewHTTP3Transport(selfcensor.SystemDialer{}, dialer.TLSDialer{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error here")
	}
	// context.Canceled error occurs if the test host supports QUIC
	// timeout error ("Handshake did not complete in time") occurs if the test host does not support QUIC
	if !(errors.Is(err, context.Canceled) || strings.HasSuffix(err.Error(), "Handshake did not complete in time")) {
		t.Fatal("not the error we expected", err)
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
}
