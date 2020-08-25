package dialer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"time"

	"github.com/ooni/probe-engine/internal/tlsx"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

// SaverDialer saves events occurring during the dial
type SaverDialer struct {
	Dialer
	Saver *trace.Saver
}

// DialContext implements Dialer.DialContext
func (d SaverDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	start := time.Now()
	conn, err := d.Dialer.DialContext(ctx, network, address)
	stop := time.Now()
	d.Saver.Write(trace.Event{
		Address:  address,
		Duration: stop.Sub(start),
		Err:      err,
		Name:     errorx.ConnectOperation,
		Proto:    network,
		Time:     stop,
	})
	return conn, err
}

// SaverTLSHandshaker saves events occurring during the handshake
type SaverTLSHandshaker struct {
	TLSHandshaker
	Saver *trace.Saver
}

// Handshake implements TLSHandshaker.Handshake
func (h SaverTLSHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	start := time.Now()
	h.Saver.Write(trace.Event{
		Name:          "tls_handshake_start",
		NoTLSVerify:   config.InsecureSkipVerify,
		TLSNextProtos: config.NextProtos,
		TLSServerName: config.ServerName,
		Time:          start,
	})
	tlsconn, state, err := h.TLSHandshaker.Handshake(ctx, conn, config)
	stop := time.Now()
	h.Saver.Write(trace.Event{
		Duration:           stop.Sub(start),
		Err:                err,
		Name:               "tls_handshake_done",
		NoTLSVerify:        config.InsecureSkipVerify,
		TLSCipherSuite:     tlsx.CipherSuiteString(state.CipherSuite),
		TLSNegotiatedProto: state.NegotiatedProtocol,
		TLSNextProtos:      config.NextProtos,
		TLSPeerCerts:       peerCerts(state, err),
		TLSServerName:      config.ServerName,
		TLSVersion:         tlsx.VersionString(state.Version),
		Time:               stop,
	})
	return tlsconn, state, err
}

// SaverConnDialer wraps the returned connection such that we
// collect all the read/write events that occur.
type SaverConnDialer struct {
	Dialer
	Saver *trace.Saver
}

// DialContext implements Dialer.DialContext
func (d SaverConnDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return saverConn{saver: d.Saver, Conn: conn}, nil
}

type saverConn struct {
	net.Conn
	saver *trace.Saver
}

func (c saverConn) Read(p []byte) (int, error) {
	start := time.Now()
	count, err := c.Conn.Read(p)
	stop := time.Now()
	c.saver.Write(trace.Event{
		Data:     p[:count],
		Duration: stop.Sub(start),
		Err:      err,
		NumBytes: count,
		Name:     errorx.ReadOperation,
		Time:     stop,
	})
	return count, err
}

func (c saverConn) Write(p []byte) (int, error) {
	start := time.Now()
	count, err := c.Conn.Write(p)
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

var _ Dialer = SaverDialer{}
var _ TLSHandshaker = SaverTLSHandshaker{}
var _ net.Conn = saverConn{}
