package dialer_test

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/dialer"
)

func TestIntegrationTLSDialerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	log.SetLevel(log.DebugLevel)
	dialer := dialer.TLSDialer{Dialer: new(net.Dialer),
		TLSHandshaker: dialer.LoggingTLSHandshaker{
			TLSHandshaker: dialer.SystemTLSHandshaker{},
			Logger:        log.Log,
		},
	}
	txp := &http.Transport{DialTLS: func(network, address string) (net.Conn, error) {
		// AlpineLinux edge is still using Go 1.13. We cannot switch to
		// using DialTLSContext here as we'd like to until either Alpine
		// switches to Go 1.14 or we drop the MK dependency.
		return dialer.DialTLSContext(context.Background(), network, address)
	}}
	client := &http.Client{Transport: txp}
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestIntegrationDNSDialerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	log.SetLevel(log.DebugLevel)
	dialer := dialer.DNSDialer{
		Dialer: dialer.LoggingDialer{
			Dialer: new(net.Dialer),
			Logger: log.Log,
		},
		Resolver: new(net.Resolver),
	}
	txp := &http.Transport{DialContext: dialer.DialContext}
	client := &http.Client{Transport: txp}
	resp, err := client.Get("http://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}
