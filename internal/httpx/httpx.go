// Package httpx contains HTTP extensions. Specifically we have code to
// create transports and clients more suitable for the OONI needs.
package httpx

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/internal/httplog"
	"github.com/ooni/probe-engine/internal/httptracex"
	"github.com/ooni/probe-engine/log"
)

// NewTransport creates a new transport suitable. The first argument is
// the function to be used to configure a network proxy. The second argument
// is the TLS client config to use. Using `nil` is fine here. Note that not
// using `nil` causes Go not to automatically upgrade to http2; see this
// issue: <https://github.com/golang/go/issues/14275>.
func NewTransport(
	proxy func(req *http.Request) (*url.URL, error), tlsConfig *tls.Config,
) *http.Transport {
	return &http.Transport{
		Proxy: proxy,
		// We use a custom dialer that retries failed
		// dialing attempts for extra robustness.
		DialContext:     (&net.Dialer{}).DialContext,
		TLSClientConfig: tlsConfig,
		// These are the same settings of Go stdlib.
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

// NewTLSConfigWithCABundle constructs a new TLS configuration using
// the specified CA bundle, or returns an error.
func NewTLSConfigWithCABundle(caBundlePath string) (*tls.Config, error) {
	cert, err := ioutil.ReadFile(caBundlePath)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(cert)
	return &tls.Config{RootCAs: pool}, nil
}

// NewTracingProxyingClient creates a new http.Client. This new http.Client
// will have the following properties:
//
// 1. it will log debug messages to the specified logger;
//
// 2. it will use netx.RetryingDialer for increased robustness;
//
// 3. it will use proxy to setup a proxy (note that passing
// nil will disable any proxy);
//
// 4. will use the specified tls.Config, if not nil (passing nil
// is preferrable; see NewTransport's docs).
func NewTracingProxyingClient(
	logger log.Logger, proxy func(req *http.Request) (*url.URL, error),
	tlsConfig *tls.Config,
) *http.Client {
	return &http.Client{
		Transport: &httptracex.Measurer{
			RoundTripper: NewTransport(proxy, tlsConfig),
			Handler: &httplog.RoundTripLogger{
				Logger: logger,
			},
		},
	}
}
