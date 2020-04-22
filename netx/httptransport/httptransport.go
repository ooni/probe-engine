// Package httptransport contains HTTP transport extensions. Here we
// define a http.Transport that emits events.
package httptransport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/resolver"
	"github.com/ooni/probe-engine/netx/trace"
)

// Dialer is the definition of dialer assumed by this package.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// TLSDialer is the definition of a TLS dialer assumed by this package.
type TLSDialer interface {
	DialTLSContext(ctx context.Context, network, address string) (net.Conn, error)
}

// RoundTripper is the definition of http.RoundTripper used by this package.
type RoundTripper interface {
	RoundTrip(req *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

// Resolver is the interface we expect from a resolver
type Resolver interface {
	LookupHost(ctx context.Context, hostname string) (addrs []string, err error)
}

// Config contains configuration for creating a new transport. When any
// field of Config is nil/empty, we will use a suitable default.
type Config struct {
	BogonIsError        bool                 // default: bogon is not error
	ByteCounter         *bytecounter.Counter // default: no explicit byte counting
	CacheResolutions    bool                 // default: no caching
	ContextByteCounting bool                 // default: no implicit byte counting
	Dialer              Dialer               // default: dialer.DNSDialer
	Logger              Logger               // default: no logging
	ProxyURL            *url.URL             // default: no proxy
	Resolver            Resolver             // default: system resolver
	Saver               *trace.Saver         // default: no saver
	TLSConfig           *tls.Config          // default: attempt using h2
	TLSDialer           TLSDialer            // default: dialer.TLSDialer
}

type tlsHandshaker interface {
	Handshake(ctx context.Context, conn net.Conn, config *tls.Config) (
		net.Conn, tls.ConnectionState, error)
}

// New creates a new RoundTripper. You can further extend the returned
// RoundTripper before wrapping it into an http.Client.
func New(config Config) RoundTripper {
	if config.Resolver == nil {
		var r Resolver = resolver.SystemResolver{}
		if config.BogonIsError {
			r = resolver.BogonResolver{Resolver: r}
		}
		r = resolver.ErrorWrapperResolver{Resolver: r}
		if config.Logger != nil {
			r = resolver.LoggingResolver{Logger: config.Logger, Resolver: r}
		}
		if config.Saver != nil {
			r = resolver.SaverResolver{Resolver: r, Saver: config.Saver}
		}
		if config.CacheResolutions {
			r = &resolver.CacheResolver{Resolver: r}
		}
		config.Resolver = r
	}
	if config.Dialer == nil {
		var d Dialer = new(net.Dialer)
		d = dialer.TimeoutDialer{Dialer: d}
		d = dialer.ErrorWrapperDialer{Dialer: d}
		if config.Logger != nil {
			d = dialer.LoggingDialer{Dialer: d, Logger: config.Logger}
		}
		if config.Saver != nil {
			d = dialer.SaverDialer{Dialer: d, Saver: config.Saver}
		}
		d = dialer.DNSDialer{Resolver: config.Resolver, Dialer: d}
		d = dialer.ProxyDialer{ProxyURL: config.ProxyURL, Dialer: d}
		if config.ContextByteCounting {
			d = dialer.ByteCounterDialer{Dialer: d}
		}
		config.Dialer = d
	}
	if config.TLSDialer == nil {
		var h tlsHandshaker = dialer.SystemTLSHandshaker{}
		h = dialer.TimeoutTLSHandshaker{TLSHandshaker: h}
		h = dialer.ErrorWrapperTLSHandshaker{TLSHandshaker: h}
		if config.Logger != nil {
			h = dialer.LoggingTLSHandshaker{Logger: config.Logger, TLSHandshaker: h}
		}
		if config.Saver != nil {
			h = dialer.SaverTLSHandshaker{TLSHandshaker: h, Saver: config.Saver}
		}
		if config.TLSConfig == nil {
			config.TLSConfig = &tls.Config{NextProtos: []string{"h2", "http/1.1"}}
		}
		config.TLSDialer = dialer.TLSDialer{
			Config:        config.TLSConfig,
			Dialer:        config.Dialer,
			TLSHandshaker: h,
		}
	}
	var txp RoundTripper
	txp = NewSystemTransport(config.Dialer, config.TLSDialer)
	if config.ByteCounter != nil {
		txp = ByteCountingTransport{Counter: config.ByteCounter, RoundTripper: txp}
	}
	if config.Logger != nil {
		txp = LoggingTransport{Logger: config.Logger, RoundTripper: txp}
	}
	if config.Saver != nil {
		txp = SaverHTTPTransport{RoundTripper: txp, Saver: config.Saver}
	}
	txp = UserAgentTransport{RoundTripper: txp}
	return txp
}
