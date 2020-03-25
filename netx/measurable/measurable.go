// Package measurable makes DNS, connect, TLS, and HTTP measurable.
package measurable

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

// Operations contains measurable operations
type Operations interface {
	LookupHost(ctx context.Context, domain string) ([]string, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	Handshake(ctx context.Context, conn net.Conn, config *tls.Config) (
		net.Conn, tls.ConnectionState, error)
	RoundTrip(req *http.Request) (*http.Response, error)
}

type operationsKey struct{}

// WithOperations returns a new context using the specified measurable operations.
func WithOperations(ctx context.Context, operations Operations) context.Context {
	return context.WithValue(ctx, operationsKey{}, operations)
}

// ContextOperations returns the Operations associated with the context.
func ContextOperations(ctx context.Context) Operations {
	operations, _ := ctx.Value(operationsKey{}).(Operations)
	return operations
}

var (
	connector     = &net.Dialer{Timeout: 30 * time.Second}
	httpTransport = &http.Transport{
		DialContext:         DialContext,
		DialTLSContext:      DialTLSContext,
		DisableCompression:  true,  // simplifies OONI measurements
		DisableKeepAlives:   false, // hey, we want to keep connections alive!
		ForceAttemptHTTP2:   true,  // we wanna use HTTP/2 if possible
		Proxy:               nil,   // use ContextDialer to implement proxying
		TLSClientConfig:     nil,   // use ContextTLSDialer instead
		TLSHandshakeTimeout: 0,     // ditto
	}
	resolver = &net.Resolver{PreferGo: false}
)

// Defaults contains the default operations implementation.
type Defaults struct{}

// LookupHost performs an host lookup
func (Defaults) LookupHost(ctx context.Context, domain string) ([]string, error) {
	return resolver.LookupHost(ctx, domain)
}

// DialContext establishes a new connection
func (Defaults) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return connector.DialContext(ctx, network, address)
}

// Handshake performs a TLS handshake
func (Defaults) Handshake(ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tlsconn := tls.Client(conn, config)
	errch := make(chan error, 1) // room for late write by goroutine
	go func() { errch <- tlsconn.Handshake() }()
	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errch:
	}
	if err != nil {
		return nil, tls.ConnectionState{}, err
	}
	return tlsconn, tlsconn.ConnectionState(), nil
}

// RoundTrip performs an HTTP round trip
func (Defaults) RoundTrip(req *http.Request) (*http.Response, error) {
	return httpTransport.RoundTrip(req)
}

// ErrConnect is the error returned when we attempted to connect multiple
// target IP addresses and all of them failed. The Errors field contains the
// result of connecting each individual target IP address.
type ErrConnect struct {
	Errors []error
}

// Error implements error.Error
func (ErrConnect) Error() string {
	return "connect_error" // compatible with ooni/probe-legacy
}

// LookupHost performs a host lookup
func LookupHost(ctx context.Context, hostname string) ([]string, error) {
	ops := ContextOperations(ctx)
	if ops == nil {
		ops = Defaults{}
	}
	return ops.LookupHost(ctx, hostname)
}

// DialContext dials a new conntection using the context's Connector and Resolver.
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	hostname, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ops := ContextOperations(ctx)
	if ops == nil {
		ops = Defaults{}
	}
	addrs, err := ops.LookupHost(ctx, hostname)
	if err != nil {
		return nil, err
	}
	if addrs == nil { // unlikely
		return nil, errors.New("defaultDialer: the resolver returned no addresses")
	}
	var errConnect ErrConnect
	for _, address = range addrs {
		address = net.JoinHostPort(address, port)
		conn, err := ops.DialContext(ctx, network, address)
		if err == nil {
			return conn, nil
		}
		errConnect.Errors = append(errConnect.Errors, err)
	}
	return nil, errConnect
}

// DialTLSContext establishes a TLS connection using the Connector, Resolver, and
// TLSHandshaker configured into the context.
func DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	return DialTLSContextConfig(ctx, network, address, nil)
}

// DialTLSContextConfig establishes a TLS connection using the Connector,
// Resolver and TLSHandshaker configured in the context, and using the provided
// config instance as a template for creating the *tls.Config. Specifically:
//
// - if config.ServerName is not set, we will extract the proper server
// name to use from the address argument
//
// - if config.NextProtos is nil, we will use "h2", "http/1.1"
//
// All the other settings inside config will not be modified. The config
// instance will however be cloned, so changes will remain local.
//
// Passing a nil config causes this function to start from a new empty config.
func DialTLSContextConfig(
	ctx context.Context, network, address string, config *tls.Config) (net.Conn, error) {
	conn, err := DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	if config == nil {
		config = new(tls.Config)
	} else {
		config = config.Clone()
	}
	if config.ServerName == "" {
		serverName, _, err := net.SplitHostPort(address)
		if err != nil { // unlikely
			conn.Close()
			return nil, err
		}
		config.ServerName = serverName
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"h2", "http/1.1"}
	}
	ops := ContextOperations(ctx)
	if ops == nil {
		ops = Defaults{}
	}
	tlsconn, _, err := ops.Handshake(ctx, conn, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return tlsconn, nil
}

type routingHTTPTransport struct{}

func (txp routingHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// make sure we see all headers that go would otherwise set
	// automatically, so that all of them can be saved
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "miniooni/0.1.0-dev")
	}
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	req.Header.Set("Host", host)
	// proceed with passing the ball to lower layers
	ops := ContextOperations(req.Context())
	if ops == nil {
		ops = Defaults{}
	}
	return ops.RoundTrip(req)
}

// DefaultHTTPTransport dispatches performing the request to the transport
// configured inside the request's context. By default, if no other transport has
// been configuered, we will use DefaultLowLevelHTTPTransport.
var DefaultHTTPTransport http.RoundTripper = routingHTTPTransport{}

// DefaultHTTPClient is the default HTTP client.
var DefaultHTTPClient = &http.Client{Transport: DefaultHTTPTransport}
