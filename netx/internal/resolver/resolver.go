// Package resolver contains code to create a resolver
package resolver

import (
	"net"
	"net/http"

	"github.com/ooni/probe-engine/netx/internal/resolver/dnstransport/dnsoverhttps"
	"github.com/ooni/probe-engine/netx/internal/resolver/dnstransport/dnsovertcp"
	"github.com/ooni/probe-engine/netx/internal/resolver/dnstransport/dnsoverudp"
	"github.com/ooni/probe-engine/netx/internal/resolver/ooniresolver"
	"github.com/ooni/probe-engine/netx/internal/resolver/parentresolver"
	"github.com/ooni/probe-engine/netx/internal/resolver/systemresolver"
	"github.com/ooni/probe-engine/netx/modelx"
)

// NewResolverSystem creates a new Go/system resolver.
func NewResolverSystem() *parentresolver.Resolver {
	return parentresolver.New(
		systemresolver.New(new(net.Resolver)),
	)
}

// NewResolverUDP creates a new UDP resolver.
func NewResolverUDP(dialer modelx.Dialer, address string) *parentresolver.Resolver {
	return parentresolver.New(
		ooniresolver.New(dnsoverudp.NewTransport(dialer, address)),
	)
}

// NewResolverTCP creates a new TCP resolver.
func NewResolverTCP(dialer modelx.Dialer, address string) *parentresolver.Resolver {
	return parentresolver.New(
		ooniresolver.New(dnsovertcp.NewTransportTCP(dialer, address)),
	)
}

// NewResolverTLS creates a new DoT resolver.
func NewResolverTLS(dialer modelx.TLSDialer, address string) *parentresolver.Resolver {
	return parentresolver.New(
		ooniresolver.New(dnsovertcp.NewTransportTLS(dialer, address)),
	)
}

// NewResolverHTTPS creates a new DoH resolver.
func NewResolverHTTPS(client *http.Client, address string) *parentresolver.Resolver {
	return parentresolver.New(
		ooniresolver.New(dnsoverhttps.NewTransport(client, address)),
	)
}
