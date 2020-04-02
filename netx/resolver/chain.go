package resolver

import (
	"context"
)

// Chain is a chain resolver.
type Chain struct {
	primary   Resolver
	secondary Resolver
}

// ChainResolvers chains two resolvers.
func ChainResolvers(primary, secondary Resolver) Chain {
	return Chain{
		primary:   primary,
		secondary: secondary,
	}
}

// LookupHost returns the IP addresses of a host
func (c Chain) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	addrs, err := c.primary.LookupHost(ctx, hostname)
	if err != nil {
		addrs, err = c.secondary.LookupHost(ctx, hostname)
	}
	return addrs, err
}
