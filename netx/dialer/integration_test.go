package dialer_test

import (
	"net"
	"net/http"
	"testing"

	"github.com/ooni/probe-engine/netx/dialer"
)

func TestIntegrationTLSDialerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	dialer := dialer.TLSDialer{Dialer: new(net.Dialer),
		TLSHandshaker: dialer.SystemTLSHandshaker{}}
	txp := &http.Transport{DialTLSContext: dialer.DialTLSContext}
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
	dialer := dialer.DNSDialer{
		Dialer:   new(net.Dialer),
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
