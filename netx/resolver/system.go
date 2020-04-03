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

var _ Resolver = SystemResolver{}
