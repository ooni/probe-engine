package quicdialer

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/legacy/netx/dialid"
	"github.com/ooni/probe-engine/netx/dialer"
)

// DNSDialer is a dialer that uses the configured Resolver to resolve a
// domain name to IP addresses
type DNSDialer struct {
	Dialer   ContextDialer
	Resolver Resolver
}

// TODO(bassosimone): figure out what `addr` is used for?

// DialContext implements ContextDialer.DialContext
func (d DNSDialer) DialContext(
	ctx context.Context, network, addr string, host string,
	tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	onlyhost, onlyport, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}
	ctx = dialid.WithDialID(ctx)
	var addrs []string
	addrs, err = d.LookupHost(ctx, onlyhost)
	if err != nil {
		return nil, err
	}
	var errorslist []error
	for _, addr := range addrs {
		target := net.JoinHostPort(addr, onlyport)
		sess, err := d.Dialer.DialContext(
			ctx, network, target, host, tlsCfg, cfg)
		if err == nil {
			return sess, nil
		}
		errorslist = append(errorslist, err)
	}
	// TODO(bassosimone): maybe ReduceErrors could be in netx/internal.
	return nil, dialer.ReduceErrors(errorslist)
}

// LookupHost implements Resolver.LookupHost
func (d DNSDialer) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if net.ParseIP(hostname) != nil {
		return []string{hostname}, nil
	}
	return d.Resolver.LookupHost(ctx, hostname)
}
