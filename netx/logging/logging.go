// Package logging adds logging to measurable objects.
package logging

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/ooni/probe-engine/netx/measurable"
)

// The Logger interface describes a generic logger. This is designed
// to be compatible with https://github.com/apex/log logger.
type Logger interface {
	Debugf(format string, v ...interface{})
	Debug(msg string)
}

// Resolver is a logging resolver.
type Resolver struct {
	measurable.Resolver
	Logger
}

// LookupHost implements measurable.Resolver.LookupHost
func (r Resolver) LookupHost(ctx context.Context, domain string) ([]string, error) {
	r.Logger.Debugf("LookupHostStart %s", domain)
	addrs, err := r.Resolver.LookupHost(ctx, domain)
	r.Logger.Debugf("LookupHostDone %+v %+v", addrs, err)
	return addrs, err
}

// Connector is a logging connector.
type Connector struct {
	measurable.Connector
	Logger
}

// DialContext implements measurable.Connector.DialContext
func (d Connector) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	d.Logger.Debugf("DialContextStart %s %s", network, address)
	conn, err := d.Connector.DialContext(ctx, network, address)
	d.Logger.Debugf("DialContextDone %+v %+v", conn, err)
	return conn, err
}

// TLSHandshaker is a logging TLS handshaker
type TLSHandshaker struct {
	measurable.TLSHandshaker
	Logger
}

// Handshake implements measurable.TLSHandshaker.Handshake
func (th TLSHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	th.Logger.Debugf("TLSHandshakeStart %+v %+v", conn, config)
	tlsconn, state, err := th.TLSHandshaker.Handshake(ctx, conn, config)
	th.Logger.Debugf("TLSHandshakeDone %+v %+v", state, err)
	return tlsconn, state, err
}

// HTTPTransport is a loggable HTTPTransport
type HTTPTransport struct {
	measurable.HTTPTransport
	Logger
}

// RoundTrip implements measurable.HTTPTransport.RoundTrip
func (txp HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	txp.Logger.Debugf("RoundTripStart %+v", req)
	resp, err := txp.HTTPTransport.RoundTrip(req)
	txp.Logger.Debugf("RoundTripDone %+v %+v", resp, err)
	return resp, err
}

// WithLogger creates a copy of the provided context that is configured
// to use the specified logger for every dial, request, etc.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	config := measurable.ContextConfigOrDefault(ctx)
	return measurable.WithConfig(ctx, &measurable.Config{
		Connector: &Connector{
			Connector: config.Connector,
			Logger:    logger,
		},
		HTTPTransport: &HTTPTransport{
			HTTPTransport: config.HTTPTransport,
			Logger:        logger,
		},
		Resolver: &Resolver{
			Resolver: config.Resolver,
			Logger:   logger,
		},
		TLSHandshaker: &TLSHandshaker{
			TLSHandshaker: config.TLSHandshaker,
			Logger:        logger,
		},
	})
}
