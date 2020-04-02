package resolver

import (
	"net"
	"net/http"

	"github.com/ooni/probe-engine/netx/modelx"
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
		NewOONIResolver(NewDNSOverUDP(dialer, address)),
	)
}

// NewResolverTCP creates a new TCP resolver.
func NewResolverTCP(dialer modelx.Dialer, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(NewDNSOverTCP(dialer, address)),
	)
}

// NewResolverTLS creates a new DoT resolver.
func NewResolverTLS(dialer modelx.TLSDialer, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(NewDNSOverTLS(dialer, address)),
	)
}

// NewResolverHTTPS creates a new DoH resolver.
func NewResolverHTTPS(client *http.Client, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(NewDNSOverHTTPS(client, address)),
	)
}
