package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"

	"github.com/lucas-clemente/quic-go"
)

// HTTP3DNSDialer is a dialer that uses the configured Resolver to resolve a
// domain name to IP addresses
type HTTP3DNSDialer struct {
	DialEarly func(net.PacketConn, net.Addr, string, *tls.Config, *quic.Config) (quic.EarlySession, error) // for testing
	Resolver  Resolver
}

// Dial implements HTTP3Dialer.Dial
func (d HTTP3DNSDialer) Dial(network, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	onlyhost, onlyport, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}
	var addrs []string
	addrs, err = d.LookupHost(onlyhost)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(onlyport)
	if err != nil {
		return nil, err
	}
	dialEarly := d.DialEarly
	if dialEarly == nil {
		dialEarly = quic.DialEarly
	}
	var errorslist []error
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		udpAddr := &net.UDPAddr{IP: ip, Port: port, Zone: ""}
		udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
		if err != nil {
			errorslist = append(errorslist, err)
			break
		}
		sess, err := dialEarly(udpConn, udpAddr, host, tlsCfg, cfg)
		if err == nil {
			return sess, nil
		}
		errorslist = append(errorslist, err)
		udpConn.Close()
	}
	return nil, reduceErrors(errorslist)
}

// LookupHost implements Resolver.LookupHost
func (d HTTP3DNSDialer) LookupHost(hostname string) ([]string, error) {
	if net.ParseIP(hostname) != nil {
		return []string{hostname}, nil
	}
	return d.Resolver.LookupHost(context.Background(), hostname)
}
