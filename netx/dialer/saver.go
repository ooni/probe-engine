package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/ooni/probe-engine/internal/tlsx"
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
		Name:     "connect",
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
	measuringconn := tlsMeasuringConn{Conn: conn, saver: h.Saver}
	proxyconn := tlsProxyConn{Conn: measuringconn}
	start := time.Now()
	h.Saver.Write(trace.Event{
		Name:          "tls_handshake_start",
		TLSNextProtos: config.NextProtos,
		TLSServerName: config.ServerName,
		Time:          start,
	})
	tlsconn, state, err := h.TLSHandshaker.Handshake(ctx, proxyconn, config)
	stop := time.Now()
	h.Saver.Write(trace.Event{
		Duration:       stop.Sub(start),
		TLSCipherSuite: tlsx.CipherSuiteString(state.CipherSuite),
		Err:            err,
		Name:           "tls_handshake_done",
		TLSNextProtos:  config.NextProtos,
		TLSPeerCerts:   state.PeerCertificates,
		Proto:          state.NegotiatedProtocol,
		TLSServerName:  config.ServerName,
		Time:           stop,
		TLSVersion:     tlsx.VersionString(state.Version),
	})
	proxyconn.Conn = conn // stop measuring
	return tlsconn, state, err
}

type tlsProxyConn struct {
	net.Conn
}

type tlsMeasuringConn struct {
	net.Conn
	saver *trace.Saver
}

func (c tlsMeasuringConn) Read(p []byte) (int, error) {
	start := time.Now()
	count, err := c.Conn.Read(p)
	stop := time.Now()
	c.saver.Write(trace.Event{
		Data:     p[:count],
		Duration: stop.Sub(start),
		Err:      err,
		NumBytes: count,
		Name:     "read",
		Time:     stop,
	})
	return count, err
}

func (c tlsMeasuringConn) Write(p []byte) (int, error) {
	start := time.Now()
	count, err := c.Conn.Write(p)
	stop := time.Now()
	c.saver.Write(trace.Event{
		Data:     p[:count],
		Duration: stop.Sub(start),
		Err:      err,
		NumBytes: count,
		Name:     "write",
		Time:     stop,
	})
	return count, err
}

var _ Dialer = SaverDialer{}
var _ TLSHandshaker = SaverTLSHandshaker{}
var _ net.Conn = tlsMeasuringConn{}
var _ net.Conn = tlsProxyConn{}
