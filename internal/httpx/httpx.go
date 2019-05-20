// Package httpx contains HTTP extensions. Specifically we have code to
// create transports and clients more suitable for the OONI needs.
package httpx

import (
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/internal/httplog"
	"github.com/ooni/probe-engine/internal/httptracex"
	"github.com/ooni/probe-engine/internal/netx"
	"github.com/ooni/probe-engine/log"
)

// NewTransport creates a new transport suitable. The argument is
// the function to be used to configure a network proxy.
func NewTransport(
	proxy func(req *http.Request) (*url.URL, error),
) *http.Transport {
	return &http.Transport{
		Proxy: proxy,
		// We use a custom dialer that retries failed
		// dialing attempts for extra robustness.
		DialContext: (&netx.RetryingDialer{}).DialContext,
		// These are the same settings of Go stdlib.
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

// NewTracingProxyingClient creates a new http.Client. This new http.Client
// will have the following properties:
//
// 1. it will log debug messages to the specified logger;
//
// 2. it will use netx.RetryingDialer for increased robustness;
//
// 3. it will use proxy to setup a proxy (note that passing
// null will disable any proxy).
func NewTracingProxyingClient(
	logger log.Logger, proxy func(req *http.Request) (*url.URL, error),
) *http.Client {
	return &http.Client{
		Transport: &httptracex.Measurer{
			RoundTripper: NewTransport(proxy),
			Handler: &httplog.RoundTripLogger{
				Logger: logger,
			},
		},
	}
}
