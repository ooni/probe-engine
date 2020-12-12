package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/internal/tlsx"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

// QUICSaverDialer saves events occurring during the dial
type QUICSaverDialer struct {
	QUICContextDialer
	Saver *trace.Saver
}

// DialContext implements Dialer.DialContext
func (d QUICSaverDialer) DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	start := time.Now()
	sess, err := d.QUICContextDialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
	stop := time.Now()
	d.Saver.Write(trace.Event{
		Address:  addr,
		Duration: stop.Sub(start),
		Err:      err,
		Name:     errorx.ConnectOperation,
		Proto:    network,
		Time:     stop,
	})
	return sess, err
}

// QUICHandshakeSaver saves events occurring during the handshake
type QUICHandshakeSaver struct {
	Saver  *trace.Saver
	Dialer QUICContextDialer
}

// DialContext implements QUICContextDialer.DialContext
func (h QUICHandshakeSaver) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	start := time.Now()
	h.Saver.Write(trace.Event{
		Name:          "tls_handshake_start",
		NoTLSVerify:   tlsCfg.InsecureSkipVerify,
		TLSNextProtos: tlsCfg.NextProtos,
		TLSServerName: tlsCfg.ServerName,
		Time:          start,
	})
	sess, err := h.Dialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
	stop := time.Now()

	if sess == nil {
		h.Saver.Write(trace.Event{
			Duration:      stop.Sub(start),
			Err:           err,
			Name:          "tls_handshake_done",
			NoTLSVerify:   tlsCfg.InsecureSkipVerify,
			TLSNextProtos: tlsCfg.NextProtos,
			TLSServerName: tlsCfg.ServerName,
			Time:          stop,
		})
		return sess, err
	}
	state := tls.ConnectionState{}
	if sess != nil {
		state = ConnectionState(sess)
	}
	h.Saver.Write(trace.Event{
		Duration:           stop.Sub(start),
		Err:                err,
		Name:               "tls_handshake_done",
		NoTLSVerify:        tlsCfg.InsecureSkipVerify,
		TLSCipherSuite:     tlsx.CipherSuiteString(state.CipherSuite),
		TLSNegotiatedProto: state.NegotiatedProtocol,
		TLSNextProtos:      tlsCfg.NextProtos,
		TLSPeerCerts:       peerCerts(state, err),
		TLSServerName:      tlsCfg.ServerName,
		TLSVersion:         tlsx.VersionString(state.Version),
		Time:               stop,
	})
	return sess, err
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

func (c saverUDPConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
	start := time.Now()
	_n, _oobn, _flags, _addr, err := c.UDPConn.ReadMsgUDP(b, oob)
	stop := time.Now()
	c.saver.Write(trace.Event{
		Data:     b[:_n],
		Duration: stop.Sub(start),
		Err:      err,
		NumBytes: _n,
		Name:     errorx.ReadOperation,
		Time:     stop,
	})
	return _n, _oobn, _flags, _addr, err
}
