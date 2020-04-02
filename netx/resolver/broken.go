package resolver

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/atomicx"
)

// BrokenResolver is a broken resolver.
type BrokenResolver struct {
	NumErrors *atomicx.Int64
}

// NewBrokenResolver creates a new broken Resolver instance.
func NewBrokenResolver() *BrokenResolver {
	return &BrokenResolver{
		NumErrors: atomicx.NewInt64(),
	}
}

var errNotFound = &net.DNSError{
	Err: "no such host",
}

// LookupAddr returns the name of the provided IP address
func (c *BrokenResolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	c.NumErrors.Add(1)
	return nil, errNotFound
}

// LookupCNAME returns the canonical name of a host
func (c *BrokenResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	c.NumErrors.Add(1)
	return "", errNotFound
}

// LookupHost returns the IP addresses of a host
func (c *BrokenResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	c.NumErrors.Add(1)
	return nil, errNotFound
}

// LookupMX returns the MX records of a specific name
func (c *BrokenResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	c.NumErrors.Add(1)
	return nil, errNotFound
}

// LookupNS returns the NS records of a specific name
func (c *BrokenResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	c.NumErrors.Add(1)
	return nil, errNotFound
}
