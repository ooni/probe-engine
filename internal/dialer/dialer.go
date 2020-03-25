// Package dialer contains the dialer implementation
package dialer

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/ooni/probe-engine/internal/errwrapper"
)

// TODO(bassosimone): here (or in tlsdialer?) we probably need support
// for wrapping the connection and saving some events.

// Dialer is the interface of all dialers
type Dialer interface {
	// DialContext is like net.Dialer.DialContext. It should split the
	// provided address using net.SplitHostPort, to get a domain name to
	// resolve. It should use some resolving functionality to map such
	// domain name to a list of IP addresses. It should then attempt to
	// dial each of them until one returns success or they all fail.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Base returns the base dialer that we use.
func Base() *net.Dialer {
	return &net.Dialer{Timeout: 30 * time.Second}
}

// Resolver is the resolver interface assumed by a dialer.
type Resolver interface {
	// LookupHost should behave like net.Resolver.LookupHost. In particular
	// it should return a single entry if hostname is an IP address.
	LookupHost(ctx context.Context, hostname string) ([]string, error)
}

// ResolvingDialer is a dialer that users a resolver.
type ResolvingDialer struct {
	Connector Dialer
	Resolver  Resolver
}

// DialContext implements Dialer.DialContext.
func (d ResolvingDialer) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	hostname, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addresses, err := d.Resolver.LookupHost(ctx, hostname)
	if err != nil {
		return nil, err
	}
	var errDial ErrDial
	for _, address := range addresses {
		address = net.JoinHostPort(address, port)
		conn, err := d.Connector.DialContext(ctx, network, address)
		if err == nil {
			return conn, nil
		}
		errDial.Errors = append(errDial.Errors, err)
	}
	return nil, errDial
}

// ErrDial indicates that DialContext failed
type ErrDial struct {
	// Errors contains the error of each connect() that failed.
	Errors []error
}

// Error implements error.Error
func (ErrDial) Error() string {
	return "connect_error"
}

// Logger is the logger interface assumed by this package
type Logger interface {
	Debugf(format string, v ...interface{})
}

// LoggingDialer is a dialer that implements logging.
type LoggingDialer struct {
	Dialer
	Logger Logger
}

// DialContext implements Dialer.DialContext.
func (d LoggingDialer) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	d.Logger.Debugf("dial %s/%s...", address, network)
	conn, err := d.Dialer.DialContext(ctx, network, address)
	d.Logger.Debugf("dial %s/%s... %+v", address, network, err)
	return conn, err
}

// ErrWrapper is a dialer that wraps errors
type ErrWrapper struct {
	Dialer
}

// DialContext implements Dialer.DialContext.
func (d ErrWrapper) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	err = errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: "connect",
	}.MaybeBuild()
	return conn, err
}

// EventsSaver is a dialer that saves events
type EventsSaver struct {
	Dialer
	events []Events
	mu     sync.Mutex
}

// ReadEvents reads the saved events and returns them
func (d *EventsSaver) ReadEvents() []Events {
	d.mu.Lock()
	ev := d.events
	d.events = nil
	d.mu.Unlock()
	return ev
}

// DialContext implements Dialer.DialContext.
func (d *EventsSaver) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	ev := Events{Network: network, RemoteAddress: address, StartTime: time.Now()}
	conn, err := d.Dialer.DialContext(ctx, network, address)
	ev.EndTime = time.Now()
	ev.Error = err
	if conn != nil {
		ev.LocalAddress = conn.LocalAddr().String()
	}
	d.mu.Lock()
	d.events = append(d.events, ev)
	d.mu.Unlock()
	return conn, err
}

// Events contains events generated when dialing
type Events struct {
	RemoteAddress string
	EndTime       time.Time
	Error         error
	LocalAddress  string
	Network       string
	StartTime     time.Time
}
