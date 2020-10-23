package httptransport

import (
	"context"
	"net"
	"net/http"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
)

// NewSystemTransport creates a new "system" HTTP transport. That is a transport
// using the Go standard library with custom dialer and TLS dialer.
func NewSystemTransport(dialer Dialer, tlsDialer TLSDialer) *http.Transport {
	txp := http.DefaultTransport.(*http.Transport).Clone()
	txp.DialContext = dialer.DialContext
	txp.DialTLSContext = tlsDialer.DialTLSContext
	// Better for Cloudflare DNS and also better because we have less
	// noisy events and we can better understand what happened.
	txp.MaxConnsPerHost = 1
	// The following (1) reduces the number of headers that Go will
	// automatically send for us and (2) ensures that we always receive
	// back the true headers, such as Content-Length. This change is
	// functional to OONI's goal of observing the network.
	txp.DisableCompression = true
	return txp
}

// HTTP3Transport consists of a http3. RoundTripper
// and fields which are not implemented by http3.Roundtripper, to mimic http.Transport
type HTTP3Transport struct {
	http3.RoundTripper
	DialContext        func(ctx context.Context, network, addr string) (net.Conn, error)
	DialTLSContext     func(ctx context.Context, network, addr string) (net.Conn, error)
	DisableCompression bool
	MaxConnsPerHost    int
}

// CloseIdleConnections TODO
func (t *HTTP3Transport) CloseIdleConnections() {
}

// NewHTTP3Transport creates a new http3 transport. That is a transport
// using the quic-go library with custom dialer and TLS dialer.
func NewHTTP3Transport(dialer Dialer, tlsDialer TLSDialer) *HTTP3Transport {
	txp := &HTTP3Transport{}
	txp.QuicConfig = new(quic.Config)
	txp.DialContext = dialer.DialContext
	txp.DialTLSContext = tlsDialer.DialTLSContext
	txp.MaxConnsPerHost = 1
	txp.DisableCompression = true
	return txp
}

var _ RoundTripper = &http.Transport{}
