package resolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitSystemResolverTransport(t *testing.T) {
	r := resolver.System{}
	transport := r.Transport()
	reply, err := transport.RoundTrip(context.Background(), nil)
	if err == nil || err.Error() != "not implemented" {
		t.Fatal("not the error we expected")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
	if transport.Address() != "" {
		t.Fatal("invalid address")
	}
	if transport.Network() != "system" {
		t.Fatal("invalid network")
	}
	if transport.RequiresPadding() != false {
		t.Fatal("we should require padding here")
	}
}

func TestIntegrationSystemResolverLookupHost(t *testing.T) {
	r := resolver.System{}
	addrs, err := r.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expected non-nil result here")
	}
}
