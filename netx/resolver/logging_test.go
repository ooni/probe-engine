package resolver_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitLoggingResolver(t *testing.T) {
	r := resolver.LoggingResolver{
		Logger:   log.Log,
		Resolver: resolver.NewFakeResolverThatFails(),
	}
	addrs, err := r.LookupHost(context.Background(), "www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if addrs != nil {
		t.Fatal("expected nil addr here")
	}
}
