package resolver

import (
	"context"
	"net/http"
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
func NewResolverUDP(dialer Dialer, address string) *ParentResolver {
	return NewParentResolver(NewOONIResolver(EmittingTransport{
		RoundTripper: NewDNSOverUDP(dialer, address),
	}))
}

// NewResolverTCP creates a new TCP resolver.
func NewResolverTCP(dial DialContextFunc, address string) *ParentResolver {
	return NewParentResolver(NewOONIResolver(EmittingTransport{
		RoundTripper: NewDNSOverTCP(dial, address),
	}))
}

// NewResolverTLS creates a new DoT resolver.
func NewResolverTLS(dial DialContextFunc, address string) *ParentResolver {
	return NewParentResolver(NewOONIResolver(EmittingTransport{
		RoundTripper: NewDNSOverTLS(dial, address),
	}))
}

// NewResolverHTTPS creates a new DoH resolver.
func NewResolverHTTPS(client *http.Client, address string) *ParentResolver {
	return NewParentResolver(
		NewOONIResolver(NewDNSOverHTTPS(client, address)),
	)
}
