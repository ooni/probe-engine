package resolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestBrokenLookupAddr(t *testing.T) {
	client := resolver.NewBrokenResolver()
	names, err := client.LookupAddr(context.Background(), "8.8.8.8")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if names != nil {
		t.Fatal("expected nil here")
	}
}

func TestBrokenLookupCNAME(t *testing.T) {
	client := resolver.NewBrokenResolver()
	cname, err := client.LookupCNAME(context.Background(), "www.ooni.io")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if cname != "" {
		t.Fatal("expected empty string here")
	}
}

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

func TestBrokenLookupMX(t *testing.T) {
	client := resolver.NewBrokenResolver()
	records, err := client.LookupMX(context.Background(), "ooni.io")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if records != nil {
		t.Fatal("expected nil here")
	}
}

func TestBrokenLookupNS(t *testing.T) {
	client := resolver.NewBrokenResolver()
	records, err := client.LookupNS(context.Background(), "ooni.io")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if records != nil {
		t.Fatal("expected nil here")
	}
}
