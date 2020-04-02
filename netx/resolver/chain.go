package resolver

import (
	"context"
)

// ChainResolver is a chain resolver.
type ChainResolver struct {
	primary   Resolver
	secondary Resolver
}

// NewChainResolver creates a new chain Resolver instance.
func NewChainResolver(primary, secondary Resolver) *ChainResolver {
	return &ChainResolver{
		primary:   primary,
		secondary: secondary,
	}
}

// LookupHost returns the IP addresses of a host
func (c *ChainResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	addrs, err := c.primary.LookupHost(ctx, hostname)
	if err != nil {
		addrs, err = c.secondary.LookupHost(ctx, hostname)
	}
	return addrs, err
}
