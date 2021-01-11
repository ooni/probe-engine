package quicdialer

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/internal/tlsx"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

// TODO(bassosimone): investigate why we have a saver for dialing
// and a saver for handshake. Not super clear currently.

// SaverDialer saves events occurring during the dial
type SaverDialer struct {
	ContextDialer
	Saver *trace.Saver
}

// DialContext implements Dialer.DialContext
func (d SaverDialer) DialContext(ctx context.Context, network, addr string,
	host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	start := time.Now()
	sess, err := d.ContextDialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
	stop := time.Now()
	d.Saver.Write(trace.Event{
		Address:  host,
		Duration: stop.Sub(start),
		Err:      err,
		Name:     errorx.QUICHandshakeOperation,
		Proto:    network,
		Time:     stop,
	})
	return sess, err
}

// HandshakeSaver saves events occurring during the handshake
type HandshakeSaver struct {
	Saver  *trace.Saver
	Dialer ContextDialer
}

// DialContext implements ContextDialer.DialContext
func (h HandshakeSaver) DialContext(ctx context.Context, network string, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	start := time.Now()
	// TODO(bassosimone): in the future we probably want to also save
	// information about what versions we're willing to accept.
	h.Saver.Write(trace.Event{
		Name:          "quic_handshake_start",
		NoTLSVerify:   tlsCfg.InsecureSkipVerify,
		TLSNextProtos: tlsCfg.NextProtos,
		TLSServerName: tlsCfg.ServerName,
		Time:          start,
	})
	sess, err := h.Dialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
	stop := time.Now()
	if err != nil {
		h.Saver.Write(trace.Event{
			Duration:      stop.Sub(start),
			Err:           err,
			Name:          "quic_handshake_done",
			NoTLSVerify:   tlsCfg.InsecureSkipVerify,
			TLSNextProtos: tlsCfg.NextProtos,
			TLSServerName: tlsCfg.ServerName,
			Time:          stop,
		})
		return nil, err
	}
	state := ConnectionState(sess)
	h.Saver.Write(trace.Event{
		Duration:           stop.Sub(start),
		Name:               "quic_handshake_done",
		NoTLSVerify:        tlsCfg.InsecureSkipVerify,
		TLSCipherSuite:     tlsx.CipherSuiteString(state.CipherSuite),
		TLSNegotiatedProto: state.NegotiatedProtocol,
		TLSNextProtos:      tlsCfg.NextProtos,
		TLSPeerCerts:       trace.PeerCerts(state, err),
		TLSServerName:      tlsCfg.ServerName,
		TLSVersion:         tlsx.VersionString(state.Version),
		Time:               stop,
	})
	return sess, err
}
