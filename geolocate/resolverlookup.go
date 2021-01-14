package geolocate

import (
	"context"
	"errors"
	"net"
)

var (
	// ErrNoIPAddressReturned indicates that no IP address was
	// returned by a specific DNS resolver.
	ErrNoIPAddressReturned = errors.New("geolocate: no IP address returned")
)

type dnsResolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
}

type resolverLookupClient struct {
	Resolver dnsResolver
}

func (rlc resolverLookupClient) Do(ctx context.Context) (string, error) {
	var ips []string
	ips, err := rlc.Resolver.LookupHost(ctx, "whoami.akamai.net")
	if err != nil {
		return "", err
	}
	if len(ips) < 1 {
		return "", ErrNoIPAddressReturned
	}
	return ips[0], nil
}

// LookupResolverIP returns the resolver IP.
func LookupResolverIP(ctx context.Context) (ip string, err error) {
	return resolverLookupClient{Resolver: &net.Resolver{}}.Do(ctx)
}
