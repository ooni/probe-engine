package resolver_test

import (
	"context"
	"net"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestChainLookupAddr(t *testing.T) {
	client := resolver.NewChainResolver(resolver.NewBrokenResolver(), new(net.Resolver))
	names, err := client.LookupAddr(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if names == nil {
		t.Fatal("expect non nil return value here")
	}
}

func TestChainLookupCNAME(t *testing.T) {
	client := resolver.NewChainResolver(resolver.NewBrokenResolver(), new(net.Resolver))
	cname, err := client.LookupCNAME(context.Background(), "www.ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if cname == "" {
		t.Fatal("expect non empty return value here")
	}
}

func TestChainLookupHost(t *testing.T) {
	client := resolver.NewChainResolver(resolver.NewBrokenResolver(), new(net.Resolver))
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expect non nil return value here")
	}
}

func TestChainLookupMX(t *testing.T) {
	client := resolver.NewChainResolver(resolver.NewBrokenResolver(), new(net.Resolver))
	records, err := client.LookupMX(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if records == nil {
		t.Fatal("expect non nil return value here")
	}
}

func TestChainLookupNS(t *testing.T) {
	client := resolver.NewChainResolver(resolver.NewBrokenResolver(), new(net.Resolver))
	records, err := client.LookupNS(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if records == nil {
		t.Fatal("expect non nil return value here")
	}
}
