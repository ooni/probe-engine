package dialer

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/legacy/netx/dialid"
)

// HTTP3DNSDialer is a dialer that uses the configured Resolver to resolve a
// domain name to IP addresses
type HTTP3DNSDialer struct {
	Dialer   HTTP3ContextDialer
	Resolver Resolver
}

// DialContext implements HTTP3Dialer.DialContext
func (d HTTP3DNSDialer) DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
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
		sess, err := d.Dialer.DialContext(ctx, network, target, host, tlsCfg, cfg)
		if err == nil {
			return sess, nil
		}
		errorslist = append(errorslist, err)
	}
	return nil, reduceErrors(errorslist)
}

// LookupHost implements Resolver.LookupHost
func (d HTTP3DNSDialer) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if net.ParseIP(hostname) != nil {
		return []string{hostname}, nil
	}
	return d.Resolver.LookupHost(ctx, hostname)
}
