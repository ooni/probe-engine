package resolver

import (
	"context"
)

// Chain is a chain resolver. The primary resolver is used first and, if that
// fails, we then attempt with the secondary resolver.
type Chain struct {
	Primary   Resolver
	Secondary Resolver
}

// LookupHost implements Resolver.LookupHost
func (c Chain) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	addrs, err := c.Primary.LookupHost(ctx, hostname)
	if err != nil {
		addrs, err = c.Secondary.LookupHost(ctx, hostname)
	}
	return addrs, err
}
