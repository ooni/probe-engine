package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/legacy/netx/dialid"
)

// HTTP3DNSDialer is a dialer that uses the configured Resolver to resolve a
// domain name to IP addresses
type HTTP3DNSDialer struct {
	DialEarlyContext func(context.Context, net.PacketConn, net.Addr, string, *tls.Config, *quic.Config) (quic.EarlySession, error) // for testing
	Resolver         Resolver
}

// DialContext implements HTTP3Dialer.DialContext
func (d HTTP3DNSDialer) DialContext(ctx context.Context, network, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
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
	port, err := strconv.Atoi(onlyport)
	if err != nil {
		return nil, err
	}
	dialEarlyContext := d.DialEarlyContext
	if dialEarlyContext == nil {
		dialEarlyContext = quic.DialEarlyContext
	}
	var errorslist []error
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		udpAddr := &net.UDPAddr{IP: ip, Port: port, Zone: ""}
		udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
		if err != nil {
			// TODO(bassosimone,kelmenhorst): we're not currently testing this
			// case, which is quite unlikely to happen, though.
			errorslist = append(errorslist, err)
			break
		}
		sess, err := dialEarlyContext(ctx, udpConn, udpAddr, host, tlsCfg, cfg)
		if err == nil {
			return sess, nil
		}
		errorslist = append(errorslist, err)
		udpConn.Close()
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
