package systemresolver

import (
	"context"
	"net"
	"testing"

	"github.com/ooni/probe-engine/netx/modelx"
)

type queryableTransport interface {
	Network() string
	Address() string
	RequiresPadding() bool
}

type queryableResolver interface {
	Transport() modelx.DNSRoundTripper
}

func TestCanQuery(t *testing.T) {
	var client modelx.DNSResolver = NewSystemResolver(new(net.Resolver))
	transport := client.(queryableResolver).Transport()
	reply, err := transport.RoundTrip(context.Background(), nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if err.Error() != "not implemented" {
		t.Fatal("not the error we expected")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
	queryableTransport := transport.(queryableTransport)
	if queryableTransport.Address() != "" {
		t.Fatal("invalid address")
	}
	if queryableTransport.Network() != "system" {
		t.Fatal("invalid network")
	}
	if queryableTransport.RequiresPadding() != false {
		t.Fatal("we should require padding here")
	}
}

func TestLookupAddr(t *testing.T) {
	client := NewSystemResolver(new(net.Resolver))
	names, err := client.LookupAddr(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if names == nil {
		t.Fatal("expected non-nil result here")
	}
}

func TestLookupCNAME(t *testing.T) {
	client := NewSystemResolver(new(net.Resolver))
	name, err := client.LookupCNAME(context.Background(), "www.ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if name == "" {
		t.Fatal("expected non-empty result here")
	}
}

func TestLookupHost(t *testing.T) {
	client := NewSystemResolver(new(net.Resolver))
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expected non-nil result here")
	}
}

func TestLookupMX(t *testing.T) {
	client := NewSystemResolver(new(net.Resolver))
	records, err := client.LookupMX(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if records == nil {
		t.Fatal("expected non-nil result here")
	}
}

func TestLookupNS(t *testing.T) {
	client := NewSystemResolver(new(net.Resolver))
	records, err := client.LookupNS(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if records == nil {
		t.Fatal("expected non-nil result here")
	}
}
