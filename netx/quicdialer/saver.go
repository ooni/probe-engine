package quicdialer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
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
		Address:  host,
		Duration: stop.Sub(start),
		Err:      err,
		Name:     errorx.QUICHandshakeOperation,
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
		Err:                err,
		Name:               "quic_handshake_done",
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

// peerCerts returns the certificates presented by the peer regardless
// of whether the TLS handshake was successful
func peerCerts(state tls.ConnectionState, err error) []*x509.Certificate {
	var x509HostnameError x509.HostnameError
	if errors.As(err, &x509HostnameError) {
		// Test case: https://wrong.host.badssl.com/
		return []*x509.Certificate{x509HostnameError.Certificate}
	}
	var x509UnknownAuthorityError x509.UnknownAuthorityError
	if errors.As(err, &x509UnknownAuthorityError) {
		// Test case: https://self-signed.badssl.com/. This error has
		// never been among the ones returned by MK.
		return []*x509.Certificate{x509UnknownAuthorityError.Cert}
	}
	var x509CertificateInvalidError x509.CertificateInvalidError
	if errors.As(err, &x509CertificateInvalidError) {
		// Test case: https://expired.badssl.com/
		return []*x509.Certificate{x509CertificateInvalidError.Cert}
	}
	return state.PeerCertificates
}
