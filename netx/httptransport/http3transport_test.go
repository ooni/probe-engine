package httptransport_test

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/selfcensor"
)

type MockHTTP3Dialer struct{}

func (d MockHTTP3Dialer) Dial(network, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	return quic.DialAddrEarly(host, tlsCfg, cfg)
}
func TestUnitHTTP3TransportSuccess(t *testing.T) {
	txp := httptransport.NewHTTP3Transport(httptransport.Config{selfcensor.SystemDialer{}, MockHTTP3Dialer{}, dialer.TLSDialer{}})

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
	txp := httptransport.NewHTTP3Transport(httptransport.Config{selfcensor.SystemDialer{}, MockHTTP3Dialer{}, dialer.TLSDialer{}})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so that the request immediately fails
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
