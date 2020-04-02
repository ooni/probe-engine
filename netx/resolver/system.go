package resolver

import (
	"context"
	"errors"

	"github.com/ooni/probe-engine/netx/modelx"
)

// SystemResolver is the system resolver
type SystemResolver struct {
	resolver modelx.DNSResolver
}

// NewSystemResolver creates a new system resolver
func NewSystemResolver(resolver modelx.DNSResolver) *SystemResolver {
	return &SystemResolver{resolver: resolver}
}

type fakeTransport struct{}

func (*fakeTransport) RoundTrip(
	ctx context.Context, query []byte,
) (reply []byte, err error) {
	return nil, errors.New("not implemented")
}

func (*fakeTransport) RequiresPadding() bool {
	return false
}

func (*fakeTransport) Network() string {
	return "system"
}

func (*fakeTransport) Address() string {
	return ""
}

// Transport returns the transport being used
func (r *SystemResolver) Transport() modelx.DNSRoundTripper {
	return &fakeTransport{}
}

// LookupHost returns the IP addresses of a host
func (r *SystemResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	return r.resolver.LookupHost(ctx, hostname)
}
