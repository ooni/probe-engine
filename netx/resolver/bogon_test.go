package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitResolverIsBogon(t *testing.T) {
	if resolver.IsBogon("antani") != true {
		t.Fatal("unexpected result")
	}
	if resolver.IsBogon("127.0.0.1") != true {
		t.Fatal("unexpected result")
	}
	if resolver.IsBogon("1.1.1.1") != false {
		t.Fatal("unexpected result")
	}
	if resolver.IsBogon("10.0.1.1") != true {
		t.Fatal("unexpected result")
	}
}

func TestUnitBogonAwareResolverWithBogon(t *testing.T) {
	r := resolver.BogonResolver{
		Resolver: resolver.NewMockableResolverWithResult([]string{"127.0.0.1"}),
	}
	addrs, err := r.LookupHost(context.Background(), "dns.google.com")
	if !errors.Is(err, modelx.ErrDNSBogon) {
		t.Fatal("not the error we expected")
	}
	if addrs != nil {
		t.Fatal("expected nil address here")
	}
}

func TestUnitBogonAwareResolverWithoutBogon(t *testing.T) {
	orig := []string{"8.8.8.8"}
	r := resolver.BogonResolver{
		Resolver: resolver.NewMockableResolverWithResult(orig),
	}
	addrs, err := r.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != len(orig) || addrs[0] != orig[0] {
		t.Fatal("not the error we expected")
	}
}
