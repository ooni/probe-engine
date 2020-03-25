// Package tlsdialer contains TLS dialers
package tlsdialer

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/ooni/probe-engine/internal/errwrapper"
	"github.com/ooni/probe-engine/internal/tlsx"
)

// Handshaker is a TLS handshaker
type Handshaker interface {
	// Handshake takes in input a plain text conn and a config and returns
	// either a TLS conn, on success, or an error, on failure. Note that this
	// method won't take ownership of the input conn. This means in particular
	// that the conn won't be closed in case of failure. The config argument
	// must not be nil, rather it should be fully initialized.
	Handshake(ctx context.Context, conn net.Conn, config *tls.Config) (
		net.Conn, tls.ConnectionState, error)
}

// StdlibHandshaker is an handshaker using the std library
type StdlibHandshaker struct{}

// Handshake implements TLSHandshaker.Handshake
func (StdlibHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	tlsconn := tls.Client(conn, config)
	errch := make(chan error, 1) // room for delayed goroutine write
	go func() {
		errch <- tlsconn.Handshake()
	}()
	select {
	case <-ctx.Done():
		return nil, tls.ConnectionState{}, ctx.Err()
	case err := <-errch:
		if err != nil {
			return nil, tls.ConnectionState{}, err
		}
		return tlsconn, tlsconn.ConnectionState(), nil
	}
}

// Logger is the logger interface assumed by this package
type Logger interface {
	Debugf(format string, v ...interface{})
}

// LoggingHandshaker is an handshaker that implements logging.
type LoggingHandshaker struct {
	Handshaker
	Logger Logger
}

// Handshake implements TLSHandshaker.Handshake
func (h LoggingHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	h.Logger.Debugf("tls %s %+v...", config.ServerName, config.NextProtos)
	tlsconn, state, err := h.Handshaker.Handshake(ctx, conn, config)
	if err != nil {
		h.Logger.Debugf("tls %s %+v... %+v", config.ServerName,
			config.NextProtos, err)
		return nil, state, err
	}
	h.Logger.Debugf("tls %s %+v... %s %s negotiated=%s",
		config.ServerName, config.NextProtos, tlsx.CipherSuiteString(state.CipherSuite),
		tlsx.VersionString(state.Version), state.NegotiatedProtocol)
	return tlsconn, state, nil
}

// ErrWrapper is an Handshaker that wraps errors
type ErrWrapper struct {
	Handshaker
}

// Handshake implements TLSHandshaker.Handshake
func (h ErrWrapper) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	tlsconn, state, err := h.Handshaker.Handshake(ctx, conn, config)
	err = errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: "tls_handshake",
	}.MaybeBuild()
	return tlsconn, state, err
}

// EventsSaver is an Handshaker that saves events
type EventsSaver struct {
	Handshaker
	events []Events
	mu     sync.Mutex
}

// ReadEvents reads the saved events and returns them
func (h *EventsSaver) ReadEvents() []Events {
	h.mu.Lock()
	ev := h.events
	h.events = nil
	h.mu.Unlock()
	return ev
}

// Handshake implements TLSHandshaker.Handshake
func (h *EventsSaver) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	ev := Events{Config: config.Clone(), StartTime: time.Now()}
	tlsconn, state, err := h.Handshaker.Handshake(ctx, conn, config)
	ev.EndTime = time.Now()
	ev.ConnState = state
	ev.Error = err
	h.mu.Lock()
	h.events = append(h.events, ev)
	h.mu.Unlock()
	return tlsconn, state, err
}

// Events contains TLS handshake events
type Events struct {
	Config    *tls.Config
	ConnState tls.ConnectionState
	EndTime   time.Time
	Error     error
	StartTime time.Time
}

// CleartextDialer is the dialer interface assumed by this package
type CleartextDialer interface {
	// DialContext is like net.Dialer.DialContext. It should split the
	// provided address using net.SplitHostPort, to get a domain name to
	// resolve. It should use some resolving functionality to map such
	// domain name to a list of IP addresses. It should then attempt to
	// dial each of them until one returns success or they all fail.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Dialer is a dialer that dials TLS connections. This interface is
// intentionally compatible with the needs of http.Transport.
type Dialer interface {
	// DialTLSContext is like Dialer.DialContext except that it will also
	// perform a TLS handshake using the host part of address.
	DialTLSContext(ctx context.Context, network, address string) (net.Conn, error)
}

// StdlibDialer is a TLS dialer using the standard library
type StdlibDialer struct {
	CleartextDialer CleartextDialer
	Handshaker      Handshaker
}

// DialContext implements Dialer.DialContext
func (d StdlibDialer) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	return d.DialTLSContext(ctx, network, address)
}

// DialTLSContext implements TLSDialer.DialTLSContext.
func (d StdlibDialer) DialTLSContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	return d.DialTLSContextWithConfig(ctx, network, address, nil)
}

// DialTLSContextWithConfig dials a TLS connection using the specified config, which
// may be nil, in which case we will use a default config. If config.NextProtos is not
// specified, then we'll advertise as next protocols "h2" and "http/1.1".
func (d StdlibDialer) DialTLSContextWithConfig(
	ctx context.Context, network, address string, config *tls.Config) (net.Conn, error) {
	hostname, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if config == nil {
		config = new(tls.Config)
	}
	config = config.Clone()
	if config.ServerName == "" {
		config.ServerName = hostname
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"h2", "http/1.1"}
	}
	conn, err := d.CleartextDialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	tlsconn, _, err := d.Handshaker.Handshake(ctx, conn, config)
	if err != nil {
		conn.Close()
		tlsconn = nil
	}
	return tlsconn, err
}
