// Package httpclient contains the default HTTP client. This client is used
// for normal operations an honours the HTTP_PROXY environment variable.
package httpclient

import (
	"net/http"

	"github.com/ooni/probe-engine/internal/dialer"
	"github.com/ooni/probe-engine/internal/httptransport"
	"github.com/ooni/probe-engine/internal/resolver"
	"github.com/ooni/probe-engine/internal/tlsdialer"
)

// Logger is the logger interface assumed by this package
type Logger interface {
	Debugf(format string, v ...interface{})
}

// New creates a new HTTP client.
func New(logger Logger) *http.Client {
	var res resolver.Resolver = resolver.Base()
	res = resolver.LoggingResolver{Resolver: res, Logger: logger}
	// TODO(bassosimone): here we can chain alternative resolvers, e.g. one that
	// performs a DoH, or DoT request with 1.1.1.1
	var dial dialer.Dialer = dialer.Base()
	dial = dialer.LoggingDialer{Dialer: dial, Logger: logger}
	dial = dialer.ResolvingDialer{Connector: dial, Resolver: res}
	// TODO(bassosimone): here we can insert into the dialing pipeline another
	// dialer that, e.g., falls back to psiphon in case of need
	var handshaker tlsdialer.Handshaker = tlsdialer.StdlibHandshaker{}
	handshaker = tlsdialer.LoggingHandshaker{Handshaker: handshaker, Logger: logger}
	var tlsdial tlsdialer.Dialer = tlsdialer.StdlibDialer{
		CleartextDialer: dial,
		Handshaker:      handshaker,
	}
	var txp httptransport.Transport = httptransport.NewBase(dial, tlsdial)
	txp = httptransport.HeaderAdder{Transport: txp, UserAgent: "miniooni/0.1.0-dev"}
	txp = httptransport.Logging{Transport: txp, Logger: logger}
	return &http.Client{Transport: txp}
}
