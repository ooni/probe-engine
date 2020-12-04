// Package httptransport contains HTTP transport extensions.
package httptransport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/lucas-clemente/quic-go"
)

// Config contains the configuration required for constructing an HTTP transport
type Config struct {
	Dialer      Dialer
	HTTP3Dialer HTTP3Dialer
	TLSDialer   TLSDialer
	TLSConfig   *tls.Config
}

// Dialer is the definition of dialer assumed by this package.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// TLSDialer is the definition of a TLS dialer assumed by this package.
type TLSDialer interface {
	DialTLSContext(ctx context.Context, network, address string) (net.Conn, error)
}

// HTTP3Dialer is the definition of dialer for HTTP3 transport assumed by this package.
type HTTP3Dialer interface {
	Dial(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

// RoundTripper is the definition of http.RoundTripper used by this package.
type RoundTripper interface {
	RoundTrip(req *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

// Resolver is the interface we expect from a resolver
type Resolver interface {
	LookupHost(ctx context.Context, hostname string) (addrs []string, err error)
	Network() string
	Address() string
}
