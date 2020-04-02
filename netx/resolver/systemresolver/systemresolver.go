// Package systemresolver contains the system resolver
package systemresolver

import (
	"context"
	"errors"
	"net"

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

// LookupAddr returns the name of the provided IP address
func (r *SystemResolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	return r.resolver.LookupAddr(ctx, addr)
}

// LookupCNAME returns the canonical name of a host
func (r *SystemResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	return r.resolver.LookupCNAME(ctx, host)
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

// LookupMX returns the MX records of a specific name
func (r *SystemResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	return r.resolver.LookupMX(ctx, name)
}

// LookupNS returns the NS records of a specific name
func (r *SystemResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	return r.resolver.LookupNS(ctx, name)
}
