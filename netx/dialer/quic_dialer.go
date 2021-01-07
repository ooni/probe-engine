package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/legacy/netx/dialid"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

// QUICContextDialer is a dialer for QUIC using Context.
type QUICContextDialer interface {
	DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

// QUICDialer is the definition of dialer for QUIC assumed by this package.
type QUICDialer interface {
	Dial(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

// QUICSystemDialer is the base dialer for QUIC
type QUICSystemDialer struct {
	Saver *trace.Saver
}

// DialContext implements QUICContextDialer.DialContext
func (d QUICSystemDialer) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	onlyhost, onlyport, err := net.SplitHostPort(addr)
	port, err := strconv.Atoi(onlyport)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(onlyhost)
	udpAddr := &net.UDPAddr{IP: ip, Port: port, Zone: ""}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	var pconn net.PacketConn = udpConn
	if d.Saver != nil {
		pconn = saverUDPConn{UDPConn: udpConn, saver: d.Saver}
	}

	sess, err := quic.DialEarlyContext(ctx, pconn, udpAddr, host, tlsCfg, cfg)
	return sess, err

}

// QUICErrorWrapperDialer is a dialer that performs quic err wrapping
type QUICErrorWrapperDialer struct {
	Dialer QUICContextDialer
}

// DialContext implements QUICContextDialer.DialContext
func (d QUICErrorWrapperDialer) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	dialID := dialid.ContextDialID(ctx)
	sess, err := d.Dialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
	err = errorx.SafeErrWrapperBuilder{
		// ConnID does not make any sense if we've failed and the error
		// does not make any sense (and is nil) if we succeded.
		DialID:    dialID,
		Error:     err,
		Operation: errorx.QUICHandshakeOperation,
		QuicErr:   true,
	}.MaybeBuild()
	if err != nil {
		return nil, err
	}
	return sess, nil
}
