// Package logging adds logging to measurable objects.
package logging

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/internal/tlsx"
	"github.com/ooni/probe-engine/netx/measurable"
)

// The Logger interface describes a generic logger. This is designed
// to be compatible with https://github.com/apex/log logger.
type Logger interface {
	Debugf(format string, v ...interface{})
	Debug(msg string)
}

// Handler adds logging to measurable operations.
type Handler struct {
	measurable.Operations
	Logger Logger
	Prefix string
}

func (lh Handler) debugf(format string, v ...interface{}) {
	if lh.Prefix != "" {
		format = lh.Prefix + " " + format
	}
	lh.Logger.Debugf(format, v...)
}

// LookupHost performs an host lookup
func (lh Handler) LookupHost(ctx context.Context, domain string) ([]string, error) {
	lh.debugf("resolve %s", domain)
	start := time.Now()
	addrs, err := lh.Operations.LookupHost(ctx, domain)
	elapsed := time.Now().Sub(start)
	lh.debugf("resolve %s => {addrs=%s err=%+v t=%s}", domain, addrs, err, elapsed)
	return addrs, err
}

// DialContext establishes a new connection
func (lh Handler) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	lh.debugf("connect %s/%s", address, network)
	start := time.Now()
	conn, err := lh.Operations.DialContext(ctx, network, address)
	elapsed := time.Now().Sub(start)
	lh.debugf("connect %s/%s => {err=%+v t=%s}", address, network, err, elapsed)
	return conn, err
}

// Handshake performs a TLS handshake
func (lh Handler) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	lh.debugf("tls {alpn=%s sni=%s}", config.NextProtos, config.ServerName)
	start := time.Now()
	tlsconn, state, err := lh.Operations.Handshake(ctx, conn, config)
	elapsed := time.Now().Sub(start)
	lh.debugf("tls {alpn=%s sni=%s} => {alpn=%s err=%+v t=%s v=%s}", config.NextProtos,
		config.ServerName, state.NegotiatedProtocol, err, elapsed,
		tlsx.VersionString(state.Version))
	return tlsconn, state, err
}

// RoundTrip performs an HTTP round trip
func (lh Handler) RoundTrip(req *http.Request) (*http.Response, error) {
	lh.debugf("%s %s", req.Method, req.URL)
	start := time.Now()
	resp, err := lh.Operations.RoundTrip(req)
	elapsed := time.Now().Sub(start)
	if err != nil {
		lh.debugf("%s %s => {err=%+v t=%s}", req.Method, req.URL, err, elapsed)
		return nil, err
	}
	lh.debugf("%s %s => {code=%+v t=%s}", req.Method, req.URL, resp.StatusCode, elapsed)
	return resp, err
}
