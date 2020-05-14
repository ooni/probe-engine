package dialer_test

import (
	"net"
	"net/http"
	"testing"

	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
)

func TestIntegration(t *testing.T) {
	txp := httptransport.New(httptransport.Config{
		Dialer: dialer.ShapingDialer{
			Dialer: new(net.Dialer),
		},
	})
	client := &http.Client{Transport: txp}
	resp, err := client.Get("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected nil response here")
	}
	resp.Body.Close()
}
