// Package logging adds logging to measurable objects.
//
// This package replaces github.com/ooni/probe-engine/internal/netxlogger as
// the code that emits logs generated through netx.
//
// Usage
//
// To enable logging, modify the context you are going to use as follows:
//
//     ctx = logging.WithLogger(ctx, logging.Config{
//         Logger: logger,
//     })
//
// where logger is a github.com/apex/log like logger. All the events generated
// using the specified context will use the configured logger.
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

// Config contains settings for the logger.
type Config struct {
	Logger Logger // logger to use
	Prefix string // optional prefix
}

type handler struct {
	measurable.Resolver
	measurable.Connector
	measurable.TLSHandshaker
	measurable.HTTPTransport
	config Config
}

func (lh handler) debugf(format string, v ...interface{}) {
	if lh.config.Prefix != "" {
		format = lh.config.Prefix + " " + format
	}
	lh.config.Logger.Debugf(format, v...)
}

func (lh handler) LookupHost(ctx context.Context, domain string) ([]string, error) {
	lh.debugf("resolve %s", domain)
	start := time.Now()
	addrs, err := lh.Resolver.LookupHost(ctx, domain)
	elapsed := time.Now().Sub(start)
	lh.debugf("resolve %s => {addrs=%s err=%+v t=%s}", domain, addrs, err, elapsed)
	return addrs, err
}

func (lh handler) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	lh.debugf("connect %s/%s", address, network)
	start := time.Now()
	conn, err := lh.Connector.DialContext(ctx, network, address)
	elapsed := time.Now().Sub(start)
	lh.debugf("connect %s/%s => {err=%+v t=%s}", address, network, err, elapsed)
	return conn, err
}

func (lh handler) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	lh.debugf("tls {alpn=%s sni=%s}", config.NextProtos, config.ServerName)
	start := time.Now()
	tlsconn, state, err := lh.TLSHandshaker.Handshake(ctx, conn, config)
	elapsed := time.Now().Sub(start)
	lh.debugf("tls {alpn=%s sni=%s} => {alpn=%s err=%+v t=%s v=%s}", config.NextProtos,
		config.ServerName, state.NegotiatedProtocol, err, elapsed,
		tlsx.VersionString(state.Version))
	return tlsconn, state, err
}

func (lh handler) RoundTrip(req *http.Request) (*http.Response, error) {
	lh.debugf("%s %s", req.Method, req.URL)
	start := time.Now()
	resp, err := lh.HTTPTransport.RoundTrip(req)
	elapsed := time.Now().Sub(start)
	if err != nil {
		lh.debugf("%s %s => {err=%+v t=%s}", req.Method, req.URL, err, elapsed)
		return nil, err
	}
	lh.debugf("%s %s => {code=%+v t=%s}", req.Method, req.URL, resp.StatusCode, elapsed)
	return resp, err
}

// WithLogger creates a copy of the provided context that is configured
// to use the specified config for every dial, request, etc.
func WithLogger(ctx context.Context, config Config) context.Context {
	cc := measurable.ContextConfigOrDefault(ctx)
	handler := handler{
		Connector:     cc.Connector,
		HTTPTransport: cc.HTTPTransport,
		Resolver:      cc.Resolver,
		TLSHandshaker: cc.TLSHandshaker,
		config:        config,
	}
	return measurable.WithConfig(ctx, &measurable.Config{
		Connector:     handler,
		HTTPTransport: handler,
		Resolver:      handler,
		TLSHandshaker: handler,
	})
}
