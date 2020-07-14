package webconnectivity_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/ooni/probe-engine/experiment/webconnectivity"
)

func TestDNSLookup(t *testing.T) {
	config := webconnectivity.DNSLookupConfig{
		Session: newsession(t, true),
		URL:     &url.URL{Host: "dns.google"},
	}
	out := webconnectivity.DNSLookup(context.Background(), config)
	if out.Failure != nil {
		t.Fatal(*out.Failure)
	}
	if len(out.Addrs) < 1 {
		t.Fatal("no addresses?!")
	}
	for addr, asn := range out.Addrs {
		if addr == "" {
			t.Fatal("empty addr")
		}
		if asn != 15169 {
			t.Fatal("invalid asn")
		}
	}
	if len(out.TestKeys.NetworkEvents) < 1 {
		t.Fatal("no network events?!")
	}
	if len(out.TestKeys.Queries) < 1 {
		t.Fatal("no queries?!")
	}
}
