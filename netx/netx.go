// Package netx contains OONI's net extensions.
//
// This package provides a replacement for net.Dialer that can Dial,
// DialContext, and DialTLS. During its lifecycle this modified Dialer
// will emit network level events using a specific emitter.
package netx

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/internal"
	"github.com/ooni/probe-engine/netx/modelx"
)

// Dialer performs measurements while dialing.
type Dialer struct {
	dialer *internal.Dialer
}

// NewDialer returns a new Dialer instance.
func NewDialer(handler modelx.Handler) *Dialer {
	return &Dialer{
		dialer: internal.NewDialer(time.Now(), handler),
	}
}

// NewDialerWithoutHandler returns a new Dialer instance.
func NewDialerWithoutHandler() *Dialer {
	return &Dialer{
		dialer: internal.NewDialer(time.Now(), handlers.NoHandler),
	}
}

// ConfigureDNS configures the DNS resolver. The network argument
// selects the type of resolver. The address argument indicates the
// resolver address and depends on the network.
//
// This functionality is not goroutine safe. You should only change
// the DNS settings before starting to use the Dialer.
//
// The following is a list of all the possible network values:
//
// - "": behaves exactly like "system"
//
// - "system": this indicates that Go should use the system resolver
// and prevents us from seeing any DNS packet. The value of the
// address parameter is ignored when using "system". If you do
// not ConfigureDNS, this is the default resolver used.
//
// - "udp": indicates that we should send queries using UDP. In this
// case the address is a host, port UDP endpoint.
//
// - "tcp": like "udp" but we use TCP.
//
// - "dot": we use DNS over TLS (DoT). In this case the address is
// the domain name of the DoT server.
//
// - "doh": we use DNS over HTTPS (DoH). In this case the address is
// the URL of the DoH server.
//
// For example:
//
//   d.ConfigureDNS("system", "")
//   d.ConfigureDNS("udp", "8.8.8.8:53")
//   d.ConfigureDNS("tcp", "8.8.8.8:53")
//   d.ConfigureDNS("dot", "dns.quad9.net")
//   d.ConfigureDNS("doh", "https://cloudflare-dns.com/dns-query")
func (d *Dialer) ConfigureDNS(network, address string) error {
	return d.dialer.ConfigureDNS(network, address)
}

// SetResolver is a more flexible way of configuring a resolver
// that should perhaps be used instead of ConfigureDNS.
func (d *Dialer) SetResolver(r modelx.DNSResolver) {
	d.dialer.SetResolver(r)
}

// Dial creates a TCP or UDP connection. See net.Dial docs.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.dialer.Dial(network, address)
}

// DialContext is like Dial but the context allows to interrupt a
// pending connection attempt at any time.
func (d *Dialer) DialContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	return d.dialer.DialContext(ctx, network, address)
}

// DialTLS is like Dial, but creates TLS connections.
func (d *Dialer) DialTLS(network, address string) (conn net.Conn, err error) {
	return d.DialTLSContext(context.Background(), network, address)
}

// DialTLSContext is like DialTLS, but with context
func (d *Dialer) DialTLSContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	return d.dialer.DialTLSContext(ctx, network, address)
}

// NewResolver returns a new resolver using the same handler of this
// Dialer. The arguments have the same meaning of ConfigureDNS. The
// returned resolver will not be used by this Dialer, and will not use
// this Dialer as well. The fact that it's a method of Dialer rather
// than an independent method is an historical oddity. There is also a
// standalone NewResolver factory and you should probably use it.
func (d *Dialer) NewResolver(network, address string) (modelx.DNSResolver, error) {
	return internal.NewResolver(d.dialer.Beginning, d.dialer.Handler, network, address)
}

// NewResolver is a standalone Dialer.NewResolver
func NewResolver(handler modelx.Handler, network, address string) (modelx.DNSResolver, error) {
	return internal.NewResolver(time.Now(), handler, network, address)
}

// NewResolverWithoutHandler creates a standalone Resolver
func NewResolverWithoutHandler(network, address string) (modelx.DNSResolver, error) {
	return internal.NewResolver(time.Now(), handlers.NoHandler, network, address)
}

// SetCABundle configures the dialer to use a specific CA bundle. This
// function is not goroutine safe. Make sure you call it before starting
// to use this specific dialer.
func (d *Dialer) SetCABundle(path string) error {
	return d.dialer.SetCABundle(path)
}

// ForceSpecificSNI forces using a specific SNI.
func (d *Dialer) ForceSpecificSNI(sni string) error {
	return d.dialer.ForceSpecificSNI(sni)
}

// ForceSkipVerify forces to skip certificate verification
func (d *Dialer) ForceSkipVerify() error {
	return d.dialer.ForceSkipVerify()
}

// ChainResolvers chains a primary and a secondary resolver such that
// we can fallback to the secondary if primary is broken.
func ChainResolvers(primary, secondary modelx.DNSResolver) modelx.DNSResolver {
	return internal.ChainResolvers(primary, secondary)
}

// HTTPTransport performs measurements during HTTP round trips.
type HTTPTransport struct {
	dialer    *internal.Dialer
	transport *internal.HTTPTransport
}

func newTransport(
	beginning time.Time, handler modelx.Handler,
	proxyFunc func(*http.Request) (*url.URL, error),
) *HTTPTransport {
	t := new(HTTPTransport)
	t.dialer = internal.NewDialer(beginning, handler)
	t.transport = internal.NewHTTPTransport(
		beginning,
		handler,
		t.dialer,
		false, // DisableKeepAlives
		proxyFunc,
	)
	return t
}

// NewHTTPTransportWithProxyFunc creates a transport without any
// handler attached using the specified proxy func.
func NewHTTPTransportWithProxyFunc(
	proxyFunc func(*http.Request) (*url.URL, error),
) *HTTPTransport {
	return newTransport(time.Now(), handlers.NoHandler, proxyFunc)
}

// NewHTTPTransport creates a new Transport. The beginning argument is
// the time to use as zero for computing the elapsed time.
func NewHTTPTransport(beginning time.Time, handler modelx.Handler) *HTTPTransport {
	return newTransport(beginning, handler, http.ProxyFromEnvironment)
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.transport.RoundTrip(req)
}

// CloseIdleConnections closes any connections which were previously connected
// from previous requests but are now sitting idle in a "keep-alive" state. It
// does not interrupt any connections currently in use.
func (t *HTTPTransport) CloseIdleConnections() {
	t.transport.CloseIdleConnections()
}

// ConfigureDNS is exactly like netx.Dialer.ConfigureDNS.
func (t *HTTPTransport) ConfigureDNS(network, address string) error {
	return t.dialer.ConfigureDNS(network, address)
}

// SetResolver is exactly like netx.Dialer.SetResolver.
func (t *HTTPTransport) SetResolver(r modelx.DNSResolver) {
	t.dialer.SetResolver(r)
}

// SetCABundle internally calls netx.Dialer.SetCABundle and
// therefore it has the same caveats and limitations.
func (t *HTTPTransport) SetCABundle(path string) error {
	return t.dialer.SetCABundle(path)
}

// ForceSpecificSNI forces using a specific SNI.
func (t *HTTPTransport) ForceSpecificSNI(sni string) error {
	return t.dialer.ForceSpecificSNI(sni)
}

// ForceSkipVerify forces to skip certificate verification
func (t *HTTPTransport) ForceSkipVerify() error {
	return t.dialer.ForceSkipVerify()
}

// HTTPClient is a replacement for http.HTTPClient.
type HTTPClient struct {
	// HTTPClient is the underlying client. Pass this client to existing code
	// that expects an *http.HTTPClient. For this reason we can't embed it.
	HTTPClient *http.Client

	// Transport is the transport configured by NewClient to be used
	// by the HTTPClient field.
	Transport *HTTPTransport
}

// NewHTTPClientWithProxyFunc creates a new client using the
// specified proxyFunc for handling proxying.
func NewHTTPClientWithProxyFunc(
	handler modelx.Handler,
	proxyFunc func(*http.Request) (*url.URL, error),
) *HTTPClient {
	transport := newTransport(time.Now(), handler, proxyFunc)
	return &HTTPClient{
		HTTPClient: &http.Client{
			Transport: transport,
		},
		Transport: transport,
	}
}

// NewHTTPClient creates a new client instance.
func NewHTTPClient(handler modelx.Handler) *HTTPClient {
	return NewHTTPClientWithProxyFunc(handler, http.ProxyFromEnvironment)
}

// NewHTTPClientWithoutProxy creates a client without any
// configured proxy attached to it. This is suitable
// for measurements where you don't want a proxy to be
// in the middle and alter the measurements.
func NewHTTPClientWithoutProxy(handler modelx.Handler) *HTTPClient {
	return NewHTTPClientWithProxyFunc(handler, nil)
}

// ConfigureDNS internally calls netx.Dialer.ConfigureDNS and
// therefore it has the same caveats and limitations.
func (c *HTTPClient) ConfigureDNS(network, address string) error {
	return c.Transport.ConfigureDNS(network, address)
}

// SetResolver internally calls netx.Dialer.SetResolver
func (c *HTTPClient) SetResolver(r modelx.DNSResolver) {
	c.Transport.SetResolver(r)
}

// SetCABundle internally calls netx.Dialer.SetCABundle and
// therefore it has the same caveats and limitations.
func (c *HTTPClient) SetCABundle(path string) error {
	return c.Transport.SetCABundle(path)
}

// ForceSpecificSNI forces using a specific SNI.
func (c *HTTPClient) ForceSpecificSNI(sni string) error {
	return c.Transport.ForceSpecificSNI(sni)
}

// ForceSkipVerify forces to skip certificate verification
func (c *HTTPClient) ForceSkipVerify() error {
	return c.Transport.ForceSkipVerify()
}
