package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/legacy/netx/dialid"
	"github.com/ooni/probe-engine/netx/errorx"
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

// QUICSystemDialer is a base dialer for QUIC
type QUICSystemDialer struct{}

// DialContext implements HTTP3ContextDialer.DialContext
func (d QUICSystemDialer) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
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

// QUICErrorWrapperDialer is a dialer that performs quic err wrapping
type QUICErrorWrapperDialer struct {
	Dialer HTTP3ContextDialer
}

// DialContext implements HTTP3ContextDialer.DialContext
func (d QUICErrorWrapperDialer) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	dialID := dialid.ContextDialID(ctx)
	sess, err := d.Dialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
	err = errorx.SafeErrWrapperBuilder{
		// ConnID does not make any sense if we've failed and the error
		// does not make any sense (and is nil) if we succeded.
		DialID:    dialID,
		Error:     err,
		Operation: errorx.ConnectOperation,
		QuicErr:   true,
	}.MaybeBuild()
	if err != nil {
		return nil, err
	}
	return sess, nil
}
