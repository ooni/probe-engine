// Package resolver contains several resolvers.
//
// TODO(bassosimone): not doing this in the PoC but we should actually
// refactor here the code that is inside netx/internal/resolver to also
// have more advanced resolving capabilities.
package resolver

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/ooni/probe-engine/internal/errwrapper"
)

// Resolver is the common interface to all resolvers.
type Resolver interface {
	// LookupHost should behave like net.Resolver.LookupHost. In particular
	// it should return a single entry if hostname is an IP address.
	LookupHost(ctx context.Context, hostname string) ([]string, error)
}

// Base returns the base resolver.
func Base() *net.Resolver {
	return &net.Resolver{PreferGo: false}
}

// Logger is the logger interface assumed by this package
type Logger interface {
	Debugf(format string, v ...interface{})
}

// LoggingResolver is a resolver that implements logging.
type LoggingResolver struct {
	Resolver
	Logger Logger
	Prefix string
}

// LookupHost implements Resolver.LookupHost
func (r LoggingResolver) LookupHost(
	ctx context.Context, hostname string) ([]string, error) {
	r.Logger.Debugf("%sresolve %s...", r.Prefix, hostname)
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	if err != nil {
		r.Logger.Debugf("%sresolve %s... %+v}", r.Prefix, hostname, err)
		return nil, err
	}
	r.Logger.Debugf("%sresolve %s... %+v", r.Prefix, hostname, addrs)
	return addrs, nil
}

// ErrWrapper is a resolver that wraps errors
type ErrWrapper struct {
	Resolver
}

// LookupHost implements Resolver.LookupHost
func (r ErrWrapper) LookupHost(
	ctx context.Context, hostname string) ([]string, error) {
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	err = errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: "resolve",
	}.MaybeBuild()
	return addrs, err
}

// EventsSaver is a resolver that saves events
type EventsSaver struct {
	Resolver
	events []Events
	mu     sync.Mutex
}

// ReadEvents reads the saved events and returns them
func (r *EventsSaver) ReadEvents() []Events {
	r.mu.Lock()
	ev := r.events
	r.events = nil
	r.mu.Unlock()
	return ev
}

// LookupHost implements Resolver.LookupHost
func (r *EventsSaver) LookupHost(
	ctx context.Context, hostname string) ([]string, error) {
	ev := Events{Hostname: hostname, StartTime: time.Now()}
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	ev.EndTime = time.Now()
	ev.Error = err
	ev.Addresses = addrs
	r.mu.Lock()
	r.events = append(r.events, ev)
	r.mu.Unlock()
	return addrs, err
}

// Events contains resolve events
type Events struct {
	Addresses []string
	EndTime   time.Time
	Error     error
	Hostname  string
	StartTime time.Time
}
