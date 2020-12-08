package dialer

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/marten-seemann/qtls-go1-15"
	"github.com/ooni/probe-engine/internal/tlsx"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

// HTTP3SaverDialer saves events occurring during the dial
type HTTP3SaverDialer struct {
	HTTP3ContextDialer
	Saver *trace.Saver
}

// DialContext implements Dialer.DialContext
func (d HTTP3SaverDialer) DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	start := time.Now()
	sess, err := d.HTTP3ContextDialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
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

// HTTP3HandshakeSaver saves events occurring during the handshake
type HTTP3HandshakeSaver struct {
	Saver  *trace.Saver
	Dialer HTTP3ContextDialer
}

// DialContext implements HTTP3ContextDialer.DialContext
func (h HTTP3HandshakeSaver) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
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
		state = qtls.ConnectionStateWith0RTT(sess.ConnectionState()).ConnectionState
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
