package resolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestBrokenLookupHost(t *testing.T) {
	client := resolver.NewBrokenResolver()
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if addrs != nil {
		t.Fatal("expected nil here")
	}
}
