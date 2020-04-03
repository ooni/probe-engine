package resolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestIntegrationSystemResolverLookupHost(t *testing.T) {
	r := resolver.SystemResolver{}
	addrs, err := r.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expected non-nil result here")
	}
}
