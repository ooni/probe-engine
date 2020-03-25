// Package measurable allows us to measure dns/dials/http requests.
//
// This package is a reimplementation of some of the functionality provided by
// github.com/ooni/probe-engine/netx after the introduction of DialTLSContext
// inside http.Transport with Go 1.14. The availability of DialTLSContext allows
// us to modify the behavior on the fly by editing the context.
//
// Accordingly, there is no logging or data collection here. Now that we can modify
// any aspect of every request with a context, we can implement different logging
// and/or data collection strategies.
//
// Thus, in terms of netx/DESIGN.md, this package only addresses the aspect of
// providing mockable/wrappable replacements for commonly used structs.
//
// Usage
//
// A design goal of this package is to make OONI code as similar as possible to
// standard library code, while still allowing measurements, logging.
//
// To dial a new network connection, your code should use:
//
//     conn, err := measurable.DialContext(ctx, network, address)
//
// Likewise, to dial a TLS connection one of:
//
//     conn, err := measurable.DialTLSContext(ctx, network, address)
//     conn, err := measurable.DialTLSContextConfig(ctx, network, address, config)
//
// To perform a name lookup:
//
//     addrs, err := measurable.LookupHost(ctx, hostname)
//
// To send an HTTP request and receive its response:
//
//     resp, err := measurable.DefaultHTTPClient.Do(req)
//
// To just perform a round trip:
//
//     resp, err := measurable.DefaultHTTPTransport.RoundTrip(req)
//
// Configuration
//
// Every dial, TLS dial, host lookup, or HTTP round trip is customizable. To do so,
// you should create and properly modify a context. Other packages should provide the
// programmer with high level functionality to implement, e.g., logging, saving the
// measurements, byte counting, retrying, and so forth.
//
// The functionality exposed by this package allows you to change the configuration
// associated with a specific context in several ways. First and foremost, you can
// get the current configuration using:
//
//     config := measurable.ContextConfig(ctx)
//     if config == nil {
//         config = measurer.NewConfig()
//     }
//
// The first time you call this function, it will return a nil Config instance
// as shown above. Once you've edited the config, you can save with:
//
//     derivedCtx := measurable.WithConfig(ctx, config)
//
// This will create a new derivedCtx context that is bound to config. You can
// access and modify the configuration stored in a ctx using:
//
//     if config := measurable.ContextConfig(ctx); config != nil {
//         config.Connector = customConnector
//     }
//
// From that point on, if config is not nil, the customConnector will be used.
package measurable

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

// The Resolver interface allows you to wrap net.Resolver as well as any
// other struct implementing a compatible interface.
//
// Possible applications:
//
// 1. log domain name resolutions
// 2. save measurement events
// 3. filter for bogons
// 4. fallback to DoH/DoT
type Resolver interface {
	// The LookupHost method discovers the IP addresses associated to
	// domain inside of the DNS. Returns a non-empty list of addresses
	// on success, and an error on failure. When domain is not actually
	// a domain name, but an IP address, this method must return a non
	// empty list containing a single entry for domain.
	LookupHost(ctx context.Context, domain string) ([]string, error)
}

// The Connector interface allows you to wrap net.Dialer as well as any
// other struct implementing a compatible interface.
//
// Possible applications:
//
// 1. log connect attempts
// 2. save measurement events
// 3. transparently implement proxying
// 4. wrap a connection
type Connector interface {
	// DialContext establishes a connection to the remote address using
	// the specified network. Returns a conn or an error.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// TLSHandshaker performs the TLS handshake.
//
// Possible applications:
//
// 1. log TLS handshake attempts
// 2. save measurement events
// 3. change params and retry handshake
// 4. use another TLS implementation
type TLSHandshaker interface {
	// Handshake performs a TLS handshake using conn and config. Returns a valid
	// conn and a valid connection state on success, an error on failure. Note that
	// this function will not close conn in case of failure.
	Handshake(ctx context.Context, conn net.Conn, config *tls.Config) (
		net.Conn, tls.ConnectionState, error)
}

// The HTTPTransport interface allows you to wrap http.Transport as well
// as any struct implementing a compatible interface. This interface is
// handy in that it has an explicit notion of CloseIdleConnections.
type HTTPTransport interface {
	// RoundTrip performs the req request and returns either the response
	// (in case of success) or an error (in case of failure). This function
	// may read part of all of the response body, as long as the body is
	// left into a state such that another transport wrapping this transport
	// may also read the body without any noticeable difference. This means
	// that, in particular, if there is an error when reading the body, also
	// the downstream transport should see the error at that point.
	RoundTrip(req *http.Request) (*http.Response, error)

	// CloseIdleConnections will close idle connections in the transport.
	CloseIdleConnections()
}

type defaultTLSHandshaker struct{}

func (defaultTLSHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
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

// Config contains the current measurable configuration. This is always
// associated with a Context, using WithConfig and ConfigContext.
type Config struct {
	Connector
	HTTPTransport
	Resolver
	TLSHandshaker
}

// NewConfig creates a new initialized Config instance.
func NewConfig() *Config {
	return &Config{
		Connector: &net.Dialer{Timeout: 30 * time.Second},
		HTTPTransport: NewLowLevelHTTPTransport(&http.Transport{
			Proxy:               nil,   // use ContextDialer to implement proxying
			TLSClientConfig:     nil,   // use ContextTLSDialer instead
			TLSHandshakeTimeout: 0,     // ditto
			DisableKeepAlives:   false, // hey, we want to keep connections alive!
			DisableCompression:  true,  // simplifies OONI measurements (but is also costly)
			ForceAttemptHTTP2:   true,  // we wanna use HTTP/2 if possible
		}),
		Resolver:      &net.Resolver{PreferGo: false},
		TLSHandshaker: defaultTLSHandshaker{},
	}
}

// configkey is the key for Context.{,With}Value.
type configkey struct{}

// WithConfig returns a copy of ctx with the specified config. It is legal
// to pass in a nil config: this means we will use default values.
func WithConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, configkey{}, config)
}

// ContextConfig returns the Config associated with the context. The return
// value of this function may be nil if no configuration has been set.
func ContextConfig(ctx context.Context) *Config {
	config, _ := ctx.Value(configkey{}).(*Config)
	return config
}

var defaultConfig *Config

func init() {
	defaultConfig = NewConfig() // necessary to break cycle
}

// ContextConfigOrDefault returns the configuration stored in the context
// or otherwise the default configuration stored in a global var.
func ContextConfigOrDefault(ctx context.Context) *Config {
	if config := ContextConfig(ctx); config != nil {
		return config
	}
	return defaultConfig
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

// LookupHost calls LookupHost on the context's Resolver.
func LookupHost(ctx context.Context, hostname string) ([]string, error) {
	config := ContextConfigOrDefault(ctx)
	return config.Resolver.LookupHost(ctx, hostname)
}

// DialContext dials a new conntection using the context's Connector and Resolver.
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	hostname, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	config := ContextConfigOrDefault(ctx)
	addrs, err := config.Resolver.LookupHost(ctx, hostname)
	if err != nil {
		return nil, err
	}
	if addrs == nil { // unlikely
		return nil, errors.New("defaultDialer: the resolver returned no addresses")
	}
	var errConnect ErrConnect
	for _, address = range addrs {
		address = net.JoinHostPort(address, port)
		conn, err := config.Connector.DialContext(ctx, network, address)
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
	ctxconf := ContextConfigOrDefault(ctx)
	tlsconn, _, err := ctxconf.TLSHandshaker.Handshake(ctx, conn, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return tlsconn, nil
}

// NewLowLevelHTTPTransport creates a low level HTTP transport that directly
// edits the provided template such that:
//
// - the DialContext member now points to DialContext
// - the DialTLSContext member now points to DialTLSContext
//
// The returned transport is a clone and can be used independently of the original.
func NewLowLevelHTTPTransport(template *http.Transport) *http.Transport {
	txp := template.Clone()
	txp.DialContext = DialContext
	txp.DialTLSContext = DialTLSContext
	return txp
}

type defaultHTTPTransport struct{}

func (txp defaultHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	config := ContextConfigOrDefault(req.Context())
	return config.HTTPTransport.RoundTrip(req)
}

// DefaultHTTPTransport dispatches performing the request to the transport
// configured inside the request's context. By default, if no other transport has
// been configuered, we will use DefaultLowLevelHTTPTransport.
var DefaultHTTPTransport http.RoundTripper = defaultHTTPTransport{}

// DefaultHTTPClient is the default HTTP client.
var DefaultHTTPClient = &http.Client{Transport: DefaultHTTPTransport}
