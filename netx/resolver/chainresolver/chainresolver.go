// Package chainresolver allows to chain two resolvers
package chainresolver

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/netx/modelx"
)

// Resolver is a chain resolver.
type Resolver struct {
	primary   modelx.DNSResolver
	secondary modelx.DNSResolver
}

// New creates a new chain Resolver instance.
func New(primary, secondary modelx.DNSResolver) *Resolver {
	return &Resolver{
		primary:   primary,
		secondary: secondary,
	}
}

// LookupAddr returns the name of the provided IP address
func (c *Resolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	names, err := c.primary.LookupAddr(ctx, addr)
	if err != nil {
		names, err = c.secondary.LookupAddr(ctx, addr)
	}
	return names, err
}

// LookupCNAME returns the canonical name of a host
func (c *Resolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	cname, err := c.primary.LookupCNAME(ctx, host)
	if err != nil {
		cname, err = c.secondary.LookupCNAME(ctx, host)
	}
	return cname, err
}

// LookupHost returns the IP addresses of a host
func (c *Resolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	addrs, err := c.primary.LookupHost(ctx, hostname)
	if err != nil {
		addrs, err = c.secondary.LookupHost(ctx, hostname)
	}
	return addrs, err
}

// LookupMX returns the MX records of a specific name
func (c *Resolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	records, err := c.primary.LookupMX(ctx, name)
	if err != nil {
		records, err = c.secondary.LookupMX(ctx, name)
	}
	return records, err
}

// LookupNS returns the NS records of a specific name
func (c *Resolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	records, err := c.primary.LookupNS(ctx, name)
	if err != nil {
		records, err = c.secondary.LookupNS(ctx, name)
	}
	return records, err
}
