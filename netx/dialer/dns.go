package dialer

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// DNSDialer defines the dialer API. We implement the most basic form
// of DNS, but more advanced resolutions are possible.
type DNSDialer struct {
	Dialer
	Resolver Resolver
}

// DialContext is like Dial but the context allows to interrupt a
// pending connection attempt at any time.
func (d DNSDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	onlyhost, onlyport, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ctx = dialid.WithDialID(ctx) // important to create before lookupHost
	var addrs []string
	addrs, err = d.lookupHost(ctx, onlyhost)
	if err != nil {
		return nil, err
	}
	var errorslist []error
	for _, addr := range addrs {
		target := net.JoinHostPort(addr, onlyport)
		conn, err := d.Dialer.DialContext(ctx, network, target)
		if err == nil {
			return conn, nil
		}
		errorslist = append(errorslist, err)
	}
	return nil, reduceErrors(errorslist)
}

func reduceErrors(errorslist []error) error {
	if len(errorslist) == 0 {
		return nil
	}
	// If we have a known error, let's consider this the real error
	// since it's probably most relevant. Otherwise let's return the
	// first considering that (1) local resolvers likely will give
	// us IPv4 first and (2) also our resolver does that. So, in case
	// the user has no IPv6 connectivity, an IPv6 error is going to
	// appear later in the list of errors.
	for _, err := range errorslist {
		var wrapper *modelx.ErrWrapper
		if errors.As(err, &wrapper) && !strings.HasPrefix(
			err.Error(), "unknown_failure",
		) {
			return err
		}
	}
	// TODO(bassosimone): handle this case in a better way
	return errorslist[0]
}

func (d DNSDialer) lookupHost(ctx context.Context, hostname string) ([]string, error) {
	if net.ParseIP(hostname) != nil {
		return []string{hostname}, nil
	}
	return d.Resolver.LookupHost(ctx, hostname)
}

// NewDNSDialer creates a new DNS dialer using the following chain:
//
// - DNSDialer (topmost)
// - EmitterDialer
// - ErrorWrapperDialer
// - TimeoutDialer
// - ByteCountingDialer
// - net.Dialer
//
// If you have others needs, manually build the chain you need.
func NewDNSDialer(resolver Resolver) DNSDialer {
	return DNSDialer{
		Dialer: EmitterDialer{
			Dialer: ErrorWrapperDialer{
				Dialer: TimeoutDialer{
					Dialer: ByteCounterDialer{
						Dialer: new(net.Dialer),
					},
				},
			},
		},
		Resolver: resolver,
	}
}
