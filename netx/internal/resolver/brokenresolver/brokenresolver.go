// Package brokenresolver is a broken resolver
package brokenresolver

import (
	"context"
	"net"
	"sync/atomic"
)

// Resolver is a broken resolver.
type Resolver struct {
	NumErrors int64
}

// New creates a new broken Resolver instance.
func New() *Resolver {
	return &Resolver{}
}

var errNotFound = &net.DNSError{
	Err: "no such host",
}

// LookupAddr returns the name of the provided IP address
func (c *Resolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	atomic.AddInt64(&c.NumErrors, 1)
	return nil, errNotFound
}

// LookupCNAME returns the canonical name of a host
func (c *Resolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	atomic.AddInt64(&c.NumErrors, 1)
	return "", errNotFound
}

// LookupHost returns the IP addresses of a host
func (c *Resolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	atomic.AddInt64(&c.NumErrors, 1)
	return nil, errNotFound
}

// LookupMX returns the MX records of a specific name
func (c *Resolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	atomic.AddInt64(&c.NumErrors, 1)
	return nil, errNotFound
}

// LookupNS returns the NS records of a specific name
func (c *Resolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	atomic.AddInt64(&c.NumErrors, 1)
	return nil, errNotFound
}
