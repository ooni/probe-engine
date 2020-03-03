package transactioner

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ooni/probe-engine/netx/internal/transactionid"
)

type transport struct {
	roundTripper http.RoundTripper
	t            *testing.T
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if id := transactionid.ContextTransactionID(ctx); id == 0 {
		t.t.Fatal("transaction ID not set")
	}
	return t.roundTripper.RoundTrip(req)
}

func TestIntegration(t *testing.T) {
	client := &http.Client{
		Transport: New(&transport{
			roundTripper: http.DefaultTransport,
			t:            t,
		}),
	}
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	client.CloseIdleConnections()
}

func TestIntegrationFailure(t *testing.T) {
	client := &http.Client{
		Transport: New(http.DefaultTransport),
	}
	// This fails the request because we attempt to speak cleartext HTTP with
	// a server that instead is expecting TLS.
	resp, err := client.Get("http://www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resp != nil {
		t.Fatal("expected a nil response here")
	}
	client.CloseIdleConnections()
}
