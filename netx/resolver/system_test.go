package resolver

import (
	"context"
	"net"
	"testing"

	"github.com/ooni/probe-engine/netx/modelx"
)

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
