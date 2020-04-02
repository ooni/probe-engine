package resolver

import (
	"net"
	"net/http"

	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/resolver/dnstransport/dnsoverhttps"
	"github.com/ooni/probe-engine/netx/resolver/dnstransport/dnsovertcp"
	"github.com/ooni/probe-engine/netx/resolver/dnstransport/dnsoverudp"
)

// NewResolverSystem creates a new Go/system resolver.
func NewResolverSystem() *ParentResolver {
	return NewParentResolver(
		NewSystemResolver(new(net.Resolver)),
	)
}

// NewResolverUDP creates a new UDP resolver.
func NewResolverUDP(dialer modelx.Dialer, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(dnsoverudp.NewDNSOverUDP(dialer, address)),
	)
}

// NewResolverTCP creates a new TCP resolver.
func NewResolverTCP(dialer modelx.Dialer, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(dnsovertcp.NewDNSOverTCP(dialer, address)),
	)
}

// NewResolverTLS creates a new DoT resolver.
func NewResolverTLS(dialer modelx.TLSDialer, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(dnsovertcp.NewDNSOverTLS(dialer, address)),
	)
}

// NewResolverHTTPS creates a new DoH resolver.
func NewResolverHTTPS(client *http.Client, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(dnsoverhttps.NewDNSOverHTTPS(client, address)),
	)
}
