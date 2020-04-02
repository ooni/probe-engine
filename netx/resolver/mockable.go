package resolver

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/atomicx"
)

// MockableResolver is a broken resolver.
type MockableResolver struct {
	NumFailures *atomicx.Int64
	err         error
	result      []string
}

// NewMockableResolverThatFails creates a new MockableResolver instance
// that always returns an error indicating NXDOMAIN.
func NewMockableResolverThatFails() MockableResolver {
	return MockableResolver{NumFailures: atomicx.NewInt64(), err: errNotFound}
}

// NewMockableResolverWithResult creates a new MockableResolver
// instance that always returns the specified result.
func NewMockableResolverWithResult(r []string) MockableResolver {
	return MockableResolver{NumFailures: atomicx.NewInt64(), result: r}
}

var errNotFound = &net.DNSError{
	Err: "no such host",
}

// LookupHost returns the IP addresses of a host
func (c MockableResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if c.err != nil {
		c.NumFailures.Add(1)
		return nil, c.err
	}
	return c.result, nil
}
