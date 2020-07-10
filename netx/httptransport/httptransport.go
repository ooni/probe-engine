// Package httptransport contains HTTP transport extensions. Here we
// define a http.Transport that emits events.
package httptransport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/certifi/gocertifi"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/resolver"
	"github.com/ooni/probe-engine/netx/selfcensor"
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
	Network() string
	Address() string
}

// Config contains configuration for creating a new transport. When any
// field of Config is nil/empty, we will use a suitable default.
//
// We use different savers for different kind of events such that the
// user of this library can choose what to save.
type Config struct {
	BaseResolver        Resolver             // default: system resolver
	BogonIsError        bool                 // default: bogon is not error
	ByteCounter         *bytecounter.Counter // default: no explicit byte counting
	CacheResolutions    bool                 // default: no caching
	ContextByteCounting bool                 // default: no implicit byte counting
	DNSCache            map[string][]string  // default: cache is empty
	DialSaver           *trace.Saver         // default: not saving dials
	Dialer              Dialer               // default: dialer.DNSDialer
	FullResolver        Resolver             // default: base resolver + goodies
	HTTPSaver           *trace.Saver         // default: not saving HTTP
	Logger              Logger               // default: no logging
	NoTLSVerify         bool                 // default: perform TLS verify
	ProxyURL            *url.URL             // default: no proxy
	ReadWriteSaver      *trace.Saver         // default: not saving read/write
	ResolveSaver        *trace.Saver         // default: not saving resolves
	TLSConfig           *tls.Config          // default: attempt using h2
	TLSDialer           TLSDialer            // default: dialer.TLSDialer
	TLSSaver            *trace.Saver         // defaukt: not saving TLS
}

type tlsHandshaker interface {
	Handshake(ctx context.Context, conn net.Conn, config *tls.Config) (
		net.Conn, tls.ConnectionState, error)
}

// CertPool is the certificate pool we're using by default
var CertPool *x509.CertPool

func init() {
	var err error
	CertPool, err = gocertifi.CACerts()
	runtimex.PanicOnError(err, "gocertifi.CACerts() failed")
}

// NewResolver creates a new resolver from the specified config
func NewResolver(config Config) Resolver {
	if config.BaseResolver == nil {
		config.BaseResolver = resolver.SystemResolver{}
	}
	var r Resolver = config.BaseResolver
	if config.CacheResolutions {
		r = &resolver.CacheResolver{Resolver: r}
	}
	if config.DNSCache != nil {
		cache := &resolver.CacheResolver{Resolver: r, ReadOnly: true}
		for key, values := range config.DNSCache {
			cache.Set(key, values)
		}
		r = cache
	}
	if config.BogonIsError {
		r = resolver.BogonResolver{Resolver: r}
	}
	r = resolver.ErrorWrapperResolver{Resolver: r}
	if config.Logger != nil {
		r = resolver.LoggingResolver{Logger: config.Logger, Resolver: r}
	}
	if config.ResolveSaver != nil {
		r = resolver.SaverResolver{Resolver: r, Saver: config.ResolveSaver}
	}
	r = resolver.AddressResolver{Resolver: r}
	return r
}

// NewDialer creates a new Dialer from the specified config
func NewDialer(config Config) Dialer {
	if config.FullResolver == nil {
		config.FullResolver = NewResolver(config)
	}
	var d Dialer = selfcensor.SystemDialer{}
	d = dialer.TimeoutDialer{Dialer: d}
	d = dialer.ErrorWrapperDialer{Dialer: d}
	if config.Logger != nil {
		d = dialer.LoggingDialer{Dialer: d, Logger: config.Logger}
	}
	if config.DialSaver != nil {
		d = dialer.SaverDialer{Dialer: d, Saver: config.DialSaver}
	}
	if config.ReadWriteSaver != nil {
		d = dialer.SaverConnDialer{Dialer: d, Saver: config.ReadWriteSaver}
	}
	d = dialer.DNSDialer{Resolver: config.FullResolver, Dialer: d}
	d = dialer.ProxyDialer{ProxyURL: config.ProxyURL, Dialer: d}
	if config.ContextByteCounting {
		d = dialer.ByteCounterDialer{Dialer: d}
	}
	d = dialer.ShapingDialer{Dialer: d}
	return d
}

// NewTLSDialer creates a new TLSDialer from the specified config
func NewTLSDialer(config Config) TLSDialer {
	if config.Dialer == nil {
		config.Dialer = NewDialer(config)
	}
	var h tlsHandshaker = dialer.SystemTLSHandshaker{}
	h = dialer.TimeoutTLSHandshaker{TLSHandshaker: h}
	h = dialer.ErrorWrapperTLSHandshaker{TLSHandshaker: h}
	if config.Logger != nil {
		h = dialer.LoggingTLSHandshaker{Logger: config.Logger, TLSHandshaker: h}
	}
	if config.TLSSaver != nil {
		h = dialer.SaverTLSHandshaker{TLSHandshaker: h, Saver: config.TLSSaver}
	}
	if config.TLSConfig == nil {
		config.TLSConfig = &tls.Config{NextProtos: []string{"h2", "http/1.1"}}
	}
	config.TLSConfig.RootCAs = CertPool // always use our own CA
	config.TLSConfig.InsecureSkipVerify = config.NoTLSVerify
	return dialer.TLSDialer{
		Config:        config.TLSConfig,
		Dialer:        config.Dialer,
		TLSHandshaker: h,
	}
}

// New creates a new RoundTripper. You can further extend the returned
// RoundTripper before wrapping it into an http.Client.
func New(config Config) RoundTripper {
	if config.Dialer == nil {
		config.Dialer = NewDialer(config)
	}
	if config.TLSDialer == nil {
		config.TLSDialer = NewTLSDialer(config)
	}
	var txp RoundTripper
	txp = NewSystemTransport(config.Dialer, config.TLSDialer)
	if config.ByteCounter != nil {
		txp = ByteCountingTransport{Counter: config.ByteCounter, RoundTripper: txp}
	}
	if config.Logger != nil {
		txp = LoggingTransport{Logger: config.Logger, RoundTripper: txp}
	}
	if config.HTTPSaver != nil {
		txp = SaverMetadataHTTPTransport{RoundTripper: txp, Saver: config.HTTPSaver}
		txp = SaverBodyHTTPTransport{RoundTripper: txp, Saver: config.HTTPSaver}
		txp = SaverPerformanceHTTPTransport{
			RoundTripper: txp, Saver: config.HTTPSaver}
		txp = SaverTransactionHTTPTransport{
			RoundTripper: txp, Saver: config.HTTPSaver}
	}
	txp = UserAgentTransport{RoundTripper: txp}
	return txp
}

// DNSClient is a DNS client. It wraps a Resolver and it possibly
// also wraps an HTTP client, but only when we're using DoH.
type DNSClient struct {
	Resolver
	httpClient *http.Client
}

// CloseIdleConnections closes idle connections, if any.
func (c DNSClient) CloseIdleConnections() {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
}

// NewDNSClient creates a new DNS client. The config argument is used to
// create the underlying Dialer and/or HTTP transport, if needed. The URL
// argument describes the kind of client that we want to make:
//
// - if the URL is `doh://powerdns`, `doh://google` or `doh://cloudflare` or the URL
// starts with `https://`, then we create a DoH client.
//
// - if the URL is `` or `system:///`, then we create a system client,
// i.e. a client using the system resolver.
//
// - if the URL starts with `udp://`, then we create a client using
// a resolver that uses the specified UDP endpoint.
//
// We return error if the URL does not parse or the URL scheme does not
// fall into one of the cases described above.
//
// If config.ResolveSaver is not nil and we're creating an underlying
// resolver where this is possible, we will also save events.
func NewDNSClient(config Config, URL string) (DNSClient, error) {
	var c DNSClient
	switch URL {
	case "doh://powerdns":
		URL = "https://doh.powerdns.org/"
	case "doh://google":
		URL = "https://dns.google/dns-query"
	case "doh://cloudflare":
		URL = "https://cloudflare-dns.com/dns-query"
	case "":
		URL = "system:///"
	}
	resolverURL, err := url.Parse(URL)
	if err != nil {
		return c, err
	}
	switch resolverURL.Scheme {
	case "system":
		c.Resolver = resolver.SystemResolver{}
		return c, nil
	case "https":
		c.httpClient = &http.Client{Transport: New(config)}
		var txp resolver.RoundTripper = resolver.NewDNSOverHTTPS(c.httpClient, URL)
		if config.ResolveSaver != nil {
			txp = resolver.SaverDNSTransport{
				RoundTripper: txp,
				Saver:        config.ResolveSaver,
			}
		}
		c.Resolver = resolver.NewSerialResolver(txp)
		return c, nil
	case "udp":
		dialer := NewDialer(config)
		var txp resolver.RoundTripper = resolver.NewDNSOverUDP(dialer, resolverURL.Host)
		if config.ResolveSaver != nil {
			txp = resolver.SaverDNSTransport{
				RoundTripper: txp,
				Saver:        config.ResolveSaver,
			}
		}
		c.Resolver = resolver.NewSerialResolver(txp)
		return c, nil
	case "dot":
		tlsDialer := NewTLSDialer(config)
		var txp resolver.RoundTripper = resolver.NewDNSOverTLS(
			tlsDialer.DialTLSContext, resolverURL.Host)
		if config.ResolveSaver != nil {
			txp = resolver.SaverDNSTransport{
				RoundTripper: txp,
				Saver:        config.ResolveSaver,
			}
		}
		c.Resolver = resolver.NewSerialResolver(txp)
		return c, nil
	case "tcp":
		dialer := NewDialer(config)
		var txp resolver.RoundTripper = resolver.NewDNSOverTCP(
			dialer.DialContext, resolverURL.Host)
		if config.ResolveSaver != nil {
			txp = resolver.SaverDNSTransport{
				RoundTripper: txp,
				Saver:        config.ResolveSaver,
			}
		}
		c.Resolver = resolver.NewSerialResolver(txp)
		return c, nil
	default:
		return c, errors.New("unsupported resolver scheme")
	}
}
