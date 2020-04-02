package resolver

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/atomicx"
)

// Mockable is a broken resolver.
type Mockable struct {
	NumFailures *atomicx.Int64
	err         error
	result      []string
}

// NewMockableResolverThatFails creates a new MockableResolver instance
// that always returns an error indicating NXDOMAIN.
func NewMockableResolverThatFails() Mockable {
	return Mockable{NumFailures: atomicx.NewInt64(), err: errNotFound}
}

// NewMockableResolverWithResult creates a new MockableResolver
// instance that always returns the specified result.
func NewMockableResolverWithResult(r []string) Mockable {
	return Mockable{NumFailures: atomicx.NewInt64(), result: r}
}

var errNotFound = &net.DNSError{
	Err: "no such host",
}

// LookupHost returns the IP addresses of a host
func (c Mockable) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if c.err != nil {
		c.NumFailures.Add(1)
		return nil, c.err
	}
	return c.result, nil
}
