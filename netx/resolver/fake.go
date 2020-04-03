package resolver

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/atomicx"
)

// FakeResolver is a mockable resolver that other packages can
// import to simulate a resolver's behaviour.
type FakeResolver struct {
	NumFailures *atomicx.Int64
	Err         error
	Result      []string
}

// NewFakeResolverThatFails creates a new MockableResolver instance
// that always returns an error indicating NXDOMAIN.
func NewFakeResolverThatFails() FakeResolver {
	return FakeResolver{NumFailures: atomicx.NewInt64(), Err: errNotFound}
}

// NewFakeResolverWithResult creates a new MockableResolver
// instance that always returns the specified result.
func NewFakeResolverWithResult(r []string) FakeResolver {
	return FakeResolver{NumFailures: atomicx.NewInt64(), Result: r}
}

var errNotFound = &net.DNSError{
	Err: "no such host",
}

// LookupHost returns the IP addresses of a host
func (c FakeResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if c.Err != nil {
		if c.NumFailures != nil {
			c.NumFailures.Add(1)
		}
		return nil, c.Err
	}
	return c.Result, nil
}

var _ Resolver = FakeResolver{}
