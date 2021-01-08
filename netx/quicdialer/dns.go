package quicdialer

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strings"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/legacy/netx/dialid"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/errorx"
)

// QUICDNSDialer is a dialer that uses the configured Resolver to resolve a
// domain name to IP addresses
type QUICDNSDialer struct {
	Dialer   QUICContextDialer
	Resolver dialer.Resolver
}

// DialContext implements QUICDialer.DialContext
func (d QUICDNSDialer) DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
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
		var wrapper *errorx.ErrWrapper
		if errors.As(err, &wrapper) && !strings.HasPrefix(
			err.Error(), "unknown_failure",
		) {
			return err
		}
	}
	// TODO(bassosimone): handle this case in a better way
	return errorslist[0]
}

// LookupHost implements Resolver.LookupHost
func (d QUICDNSDialer) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if net.ParseIP(hostname) != nil {
		return []string{hostname}, nil
	}
	return d.Resolver.LookupHost(ctx, hostname)
}
