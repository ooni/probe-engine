package resolver

import (
	"context"
	"net"
)

// SystemResolver is the system resolver
type SystemResolver struct{}

// LookupHost returns the IP addresses of a host
func (r SystemResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	return net.DefaultResolver.LookupHost(ctx, hostname)
}

// Network implements Resolver.Network
func (r SystemResolver) Network() string {
	return "system"
}

// Address implements Resolver.Address
func (r SystemResolver) Address() string {
	return ""
}

var _ Resolver = SystemResolver{}
