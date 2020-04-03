package resolver

import (
	"context"
	"errors"
	"net"
)

// SystemResolver is the system resolver
type SystemResolver struct{}

// SystemTransport is the fake transport for the system resolver
type SystemTransport struct{}

// RoundTrip implements RoundTripper.RoundTrip
func (SystemTransport) RoundTrip(ctx context.Context, query []byte) (reply []byte, err error) {
	return nil, errors.New("not implemented")
}

// RequiresPadding implements RoundTripper.RequiresPadding
func (SystemTransport) RequiresPadding() bool {
	return false
}

// Network implements RoundTripper.Network
func (SystemTransport) Network() string {
	return "system"
}

// Address implements RoundTripper.Address
func (SystemTransport) Address() string {
	return ""
}

// Transport returns the transport being used
func (r SystemResolver) Transport() RoundTripper {
	return SystemTransport{}
}

// LookupHost returns the IP addresses of a host
func (r SystemResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	return net.DefaultResolver.LookupHost(ctx, hostname)
}

var _ Resolver = SystemResolver{}
var _ RoundTripper = SystemTransport{}
