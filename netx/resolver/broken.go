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
	return &BrokenResolver{NumErrors: atomicx.NewInt64()}
}

var errNotFound = &net.DNSError{
	Err: "no such host",
}

// LookupHost returns the IP addresses of a host
func (c *BrokenResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	c.NumErrors.Add(1)
	return nil, errNotFound
}
