package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	"github.com/lucas-clemente/quic-go"
)

// QUICBaseDialer is a system dialer for Quic
type QUICBaseDialer interface {
	DialEarlyContext(context.Context, net.PacketConn, net.Addr, string, *tls.Config, *quic.Config) (quic.EarlySession, error)
}

// HTTP3ContextDialer is a dialer for HTTP3 transport using Context.
type HTTP3ContextDialer interface {
	DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

// HTTP3Dialer is the definition of dialer for HTTP3 transport assumed by this package.
type HTTP3Dialer interface {
	Dial(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

type SystemBaseDialer struct{}

func (d SystemBaseDialer) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	onlyhost, onlyport, err := net.SplitHostPort(addr)
	port, err := strconv.Atoi(onlyport)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(onlyhost)
	udpAddr := &net.UDPAddr{IP: ip, Port: port, Zone: ""}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})

	sess, err := quic.DialEarlyContext(ctx, udpConn, udpAddr, host, tlsCfg, cfg)
	go func() {
		// wait before closing the connection
		time.Sleep(2 * time.Second)
		udpConn.Close()
	}()
	return sess, err
}
