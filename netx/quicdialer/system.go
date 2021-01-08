package quicdialer

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

// QUICSystemDialer is the basic dialer for QUIC
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
	if ip == nil {
		return nil, errors.New("invalid IP representation")
	}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	var pconn net.PacketConn = udpConn
	if d.Saver != nil {
		pconn = saverUDPConn{UDPConn: udpConn, saver: d.Saver}
	}

	udpAddr := &net.UDPAddr{IP: ip, Port: port, Zone: ""}
	return quic.DialEarlyContext(ctx, pconn, udpAddr, host, tlsCfg, cfg)

}

// saverUDPConn is used by the QUIC System Dialer if a ReadWriteSaver is set in the netx config
type saverUDPConn struct {
	*net.UDPConn
	saver *trace.Saver
}

func (c saverUDPConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	start := time.Now()
	count, err := c.UDPConn.WriteTo(p, addr)
	stop := time.Now()
	c.saver.Write(trace.Event{
		Data:     p[:count],
		Duration: stop.Sub(start),
		Err:      err,
		NumBytes: count,
		Name:     errorx.WriteOperation,
		Time:     stop,
	})
	return count, err
}

func (c saverUDPConn) ReadMsgUDP(b, oob []byte) (int, int, int, *net.UDPAddr, error) {
	start := time.Now()
	n, oobn, flags, addr, err := c.UDPConn.ReadMsgUDP(b, oob)
	var data []byte
	if n > 0 {
		data = b[:n]
	}
	stop := time.Now()
	c.saver.Write(trace.Event{
		Data:     data,
		Duration: stop.Sub(start),
		Err:      err,
		NumBytes: n,
		Name:     errorx.ReadOperation,
		Time:     stop,
	})
	return n, oobn, flags, addr, err
}
