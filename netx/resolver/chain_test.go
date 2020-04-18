package resolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestChainLookupHost(t *testing.T) {
	client := resolver.ChainResolver{
		Primary:   resolver.NewFakeResolverThatFails(),
		Secondary: resolver.SystemResolver{},
	}
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expect non nil return value here")
	}
}
