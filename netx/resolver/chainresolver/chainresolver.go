// Package chainresolver allows to chain two resolvers
package chainresolver

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/netx/modelx"
)

// ChainResolver is a chain resolver.
type ChainResolver struct {
	primary   modelx.DNSResolver
	secondary modelx.DNSResolver
}

// NewChainResolver creates a new chain Resolver instance.
func NewChainResolver(primary, secondary modelx.DNSResolver) *ChainResolver {
	return &ChainResolver{
		primary:   primary,
		secondary: secondary,
	}
}

// LookupAddr returns the name of the provided IP address
func (c *ChainResolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	names, err := c.primary.LookupAddr(ctx, addr)
	if err != nil {
		names, err = c.secondary.LookupAddr(ctx, addr)
	}
	return names, err
}

// LookupCNAME returns the canonical name of a host
func (c *ChainResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	cname, err := c.primary.LookupCNAME(ctx, host)
	if err != nil {
		cname, err = c.secondary.LookupCNAME(ctx, host)
	}
	return cname, err
}

// LookupHost returns the IP addresses of a host
func (c *ChainResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	addrs, err := c.primary.LookupHost(ctx, hostname)
	if err != nil {
		addrs, err = c.secondary.LookupHost(ctx, hostname)
	}
	return addrs, err
}

// LookupMX returns the MX records of a specific name
func (c *ChainResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	records, err := c.primary.LookupMX(ctx, name)
	if err != nil {
		records, err = c.secondary.LookupMX(ctx, name)
	}
	return records, err
}

// LookupNS returns the NS records of a specific name
func (c *ChainResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	records, err := c.primary.LookupNS(ctx, name)
	if err != nil {
		records, err = c.secondary.LookupNS(ctx, name)
	}
	return records, err
}
