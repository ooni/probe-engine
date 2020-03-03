// Package systemresolver contains the system resolver
package systemresolver

import (
	"context"
	"errors"
	"net"

	"github.com/ooni/probe-engine/netx/modelx"
)

// Resolver is the system resolver
type Resolver struct {
	resolver modelx.DNSResolver
}

// New creates a new system resolver
func New(resolver modelx.DNSResolver) *Resolver {
	return &Resolver{resolver: resolver}
}

// LookupAddr returns the name of the provided IP address
func (r *Resolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	return r.resolver.LookupAddr(ctx, addr)
}

// LookupCNAME returns the canonical name of a host
func (r *Resolver) LookupCNAME(ctx context.Context, host string) (string, error) {
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
func (r *Resolver) Transport() modelx.DNSRoundTripper {
	return &fakeTransport{}
}

// LookupHost returns the IP addresses of a host
func (r *Resolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	return r.resolver.LookupHost(ctx, hostname)
}

// LookupMX returns the MX records of a specific name
func (r *Resolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	return r.resolver.LookupMX(ctx, name)
}

// LookupNS returns the NS records of a specific name
func (r *Resolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	return r.resolver.LookupNS(ctx, name)
}
