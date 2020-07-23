// Package uncensored contains uncensored facilities. These facilities
// are used by Jafar code to evade its own censorship efforts.
package uncensored

import (
	"context"
	"net"
	"net/http"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/netx/httptransport"
)

// Client is DNS, HTTP, and TCP client.
type Client struct {
	dnsClient     *httptransport.DNSClient
	httpTransport httptransport.RoundTripper
	dialer        httptransport.Dialer
}

// NewClient creates a new Client.
func NewClient(resolverURL string) (*Client, error) {
	configuration, err := urlgetter.Configurer{
		Config: urlgetter.Config{
			ResolverURL: resolverURL,
		},
		Logger: log.Log,
	}.NewConfiguration()
	if err != nil {
		return nil, err
	}
	return &Client{
		dnsClient:     &configuration.DNSClient,
		httpTransport: httptransport.New(configuration.HTTPConfig),
		dialer:        httptransport.NewDialer(configuration.HTTPConfig),
	}, nil
}

// Must panics if it's not possible to create a Client. Usually you should
// use it like `uncensored.Must(uncensored.NewClient(URL))`.
func Must(client *Client, err error) *Client {
	runtimex.PanicOnError(err, "cannot create uncensored client")
	return client
}

// DefaultClient is the default client for DNS, HTTP, and TCP.
var DefaultClient = Must(NewClient(""))

var _ httptransport.Resolver = DefaultClient

// Address implements httptransport.Resolver.Address
func (c *Client) Address() string {
	return c.dnsClient.Address()
}

// LookupHost implements httptransport.Resolver.LookupHost
func (c *Client) LookupHost(ctx context.Context, domain string) ([]string, error) {
	return c.dnsClient.LookupHost(ctx, domain)
}

// Network implements httptransport.Resolver.Network
func (c *Client) Network() string {
	return c.dnsClient.Network()
}

var _ httptransport.Dialer = DefaultClient

// DialContext implements httptransport.Dialer.DialContext
func (c *Client) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return c.dialer.DialContext(ctx, network, address)
}

var _ httptransport.RoundTripper = DefaultClient

// CloseIdleConnections implement httptransport.RoundTripper.CloseIdleConnections
func (c *Client) CloseIdleConnections() {
	c.dnsClient.CloseIdleConnections()
	c.httpTransport.CloseIdleConnections()
}

// RoundTrip implement httptransport.RoundTripper.RoundTrip
func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.httpTransport.RoundTrip(req)
}
