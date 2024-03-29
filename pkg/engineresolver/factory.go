package engineresolver

import (
	"errors"
	"net/url"

	"github.com/ooni/probe-engine/pkg/bytecounter"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// errCannotUseHTTP3WithAProxyURL means we cannot construct a new
// child resolver using HTTP/3 with a proxy URL.
var errCannotUseHTTP3WithAProxyURL = errors.New("cannot use HTTP/3 with a proxy URL")

// errUnsupportedResolverScheme means we don't support the
// given resolver scheme. We only support https, http and system.
var errUnsupportedResolverScheme = errors.New("unsupported resolver scheme")

// newChildResolver constructs a new child resolver.
//
// Arguments:
//
// - logger is the MANDATORY logger;
//
// - URL is the MANDATORY URL to use (a DoH URL or system:///);
//
// - http3Enabled indicates whether to use HTTP/3;
//
// - counter is the OPTIONAL byte counter;
//
// - proxyURL is the OPTIONAL proxy URL.
//
// Using a proxy URL is incompatible with using HTTP/3 and this
// factory will return an error if that happens.
//
// This function returns a model.Resolver or an error.
func newChildResolver(
	logger model.Logger,
	URL string,
	http3Enabled bool,
	counter *bytecounter.Counter,
	proxyURL *url.URL,
) (model.Resolver, error) {
	runtimex.Assert(logger != nil, "passed a nil model.Logger")
	runtimex.Assert(URL != "", "passed an empty URL")
	if http3Enabled && proxyURL != nil {
		return nil, errCannotUseHTTP3WithAProxyURL
	}
	parsed, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	var reso model.Resolver
	switch parsed.Scheme {
	case "http", "https": // http is here for testing
		reso = newChildResolverHTTPS(logger, URL, http3Enabled, counter, proxyURL)
	case "system":
		netx := &netxlite.Netx{}
		reso = bytecounter.MaybeWrapSystemResolver(
			netx.NewStdlibResolver(logger),
			counter, // handles correctly the case where counter is nil
		)
	default:
		return nil, errUnsupportedResolverScheme
	}
	reso = netxlite.MaybeWrapWithBogonResolver(true, reso)
	return reso, nil
}

// newChildResolverHTTPS is like newChildResolver but assumes that
// we already know that the URL scheme is http or https.
func newChildResolverHTTPS(
	logger model.Logger,
	URL string,
	http3Enabled bool,
	counter *bytecounter.Counter,
	proxyURL *url.URL,
) model.Resolver {
	var txp model.HTTPTransport
	netx := &netxlite.Netx{}
	switch http3Enabled {
	case false:
		dialer := netxlite.NewDialerWithStdlibResolver(logger)
		thx := netx.NewTLSHandshakerStdlib(logger)
		tlsDialer := netxlite.NewTLSDialer(dialer, thx)
		txp = netxlite.NewHTTPTransportWithOptions(
			logger, dialer, tlsDialer,
			netxlite.HTTPTransportOptionDisableCompression(false), // defaults to true but compression is fine here
			netxlite.HTTPTransportOptionProxyURL(proxyURL),        // nil here disables using the proxy
		)
	case true:
		txp = netx.NewHTTP3TransportStdlib(logger)
	}
	txp = bytecounter.MaybeWrapHTTPTransport(txp, counter)
	dnstxp := netxlite.NewDNSOverHTTPSTransportWithHTTPTransport(txp, URL)
	underlying := netxlite.NewUnwrappedParallelResolver(dnstxp)
	wrapped := netxlite.WrapResolver(logger, underlying)
	return wrapped
}
