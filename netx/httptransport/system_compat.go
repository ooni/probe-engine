// +build !go1.14

package httptransport

import (
	"context"
	"net"
	"net/http"
)

// NewSystemTransport creates a new "system" HTTP transport. That is a transport
// using the Go standard library with custom dialer and TLS dialer.
func NewSystemTransport(dialer Dialer, tlsDialer TLSDialer, proxy ProxyFunc) *http.Transport {
	txp := http.DefaultTransport.(*http.Transport).Clone()
	txp.Proxy = proxy
	txp.DialContext = dialer.DialContext
	txp.DialTLS = func(network, address string) (net.Conn, error) {
		// Go < 1.14 does not have http.Transport.DialTLSContext
		return tlsDialer.DialTLSContext(context.Background(), network, address)
	}
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

var _ RoundTripper = &http.Transport{}
