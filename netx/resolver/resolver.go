package resolver

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/netx/modelx"
)

// Resolver is a DNS resolver. The *net.Resolver used by Go implements
// this interface, but other implementations are possible.
type Resolver interface {
	// LookupHost resolves a hostname to a list of IP addresses.
	LookupHost(ctx context.Context, hostname string) (addrs []string, err error)
}

// NewResolverSystem creates a new Go/system resolver.
func NewResolverSystem() *ParentResolver {
	return NewParentResolver(System{})
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
