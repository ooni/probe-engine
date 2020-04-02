package resolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitMockableResolverThatFails(t *testing.T) {
	client := resolver.NewMockableResolverThatFails()
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if addrs != nil {
		t.Fatal("expected nil here")
	}
}

func TestUnitMockableResolverWithResult(t *testing.T) {
	orig := []string{"10.0.0.1"}
	client := resolver.NewMockableResolverWithResult(orig)
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(orig) != len(addrs) || orig[0] != addrs[0] {
		t.Fatal("not the result we expected")
	}
}
