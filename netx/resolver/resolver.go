// Package resolver contains code to create a resolver
package resolver

import (
	"net"
	"net/http"

	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/resolver/dnstransport/dnsoverhttps"
	"github.com/ooni/probe-engine/netx/resolver/dnstransport/dnsovertcp"
	"github.com/ooni/probe-engine/netx/resolver/dnstransport/dnsoverudp"
	"github.com/ooni/probe-engine/netx/resolver/ooniresolver"
	"github.com/ooni/probe-engine/netx/resolver/parentresolver"
	"github.com/ooni/probe-engine/netx/resolver/systemresolver"
)

// NewResolverSystem creates a new Go/system resolver.
func NewResolverSystem() *parentresolver.ParentResolver {
	return parentresolver.NewParentResolver(
		systemresolver.NewSystemResolver(new(net.Resolver)),
	)
}

// NewResolverUDP creates a new UDP resolver.
func NewResolverUDP(dialer modelx.Dialer, address string) *parentresolver.ParentResolver {
	return parentresolver.NewParentResolver(
		ooniresolver.NewOONIResolver(dnsoverudp.NewTransport(dialer, address)),
	)
}

// NewResolverTCP creates a new TCP resolver.
func NewResolverTCP(dialer modelx.Dialer, address string) *parentresolver.ParentResolver {
	return parentresolver.NewParentResolver(
		ooniresolver.NewOONIResolver(dnsovertcp.NewTransportTCP(dialer, address)),
	)
}

// NewResolverTLS creates a new DoT resolver.
func NewResolverTLS(dialer modelx.TLSDialer, address string) *parentresolver.ParentResolver {
	return parentresolver.NewParentResolver(
		ooniresolver.NewOONIResolver(dnsovertcp.NewTransportTLS(dialer, address)),
	)
}

// NewResolverHTTPS creates a new DoH resolver.
func NewResolverHTTPS(client *http.Client, address string) *parentresolver.ParentResolver {
	return parentresolver.NewParentResolver(
		ooniresolver.NewOONIResolver(dnsoverhttps.NewTransport(client, address)),
	)
}
