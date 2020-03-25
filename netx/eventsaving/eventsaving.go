// Package eventsaving uses measurable to save events.
package eventsaving

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ooni/probe-engine/netx/measurable"
)

// An Event describes something that occurred.
type Event struct {
	Op   string
	Time time.Time
}

// The Saver struct describes a generic events saver.
type Saver struct {
	mu     sync.Mutex
	events []Event
}

func (s *Saver) emit(ev Event) {
	ev.Time = time.Now()
	s.mu.Lock()
	s.events = append(s.events, ev)
	s.mu.Unlock()
}

// ReadEvents removes the bufferized events from the internal
// queue and returns them to the caller.
func (s *Saver) ReadEvents() []Event {
	s.mu.Lock()
	evs := s.events
	s.events = nil
	s.mu.Unlock()
	return evs
}

// Resolver is an event saving resolver.
type Resolver struct {
	measurable.Resolver
	*Saver
}

// LookupHost implements measurable.Resolver.LookupHost
func (r Resolver) LookupHost(ctx context.Context, domain string) ([]string, error) {
	r.Saver.emit(Event{Op: "LookupHostStart"})
	addrs, err := r.Resolver.LookupHost(ctx, domain)
	r.Saver.emit(Event{Op: "LookupHostDone"})
	return addrs, err
}

// Connector is an event saving connector.
type Connector struct {
	measurable.Connector
	*Saver
}

// DialContext implements measurable.Connector.DialContext
func (d Connector) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	d.Saver.emit(Event{Op: "DialContextStart"})
	conn, err := d.Connector.DialContext(ctx, network, address)
	d.Saver.emit(Event{Op: "DialContextDone"})
	return conn, err
}

// TLSHandshaker is an event saving TLS handshaker
type TLSHandshaker struct {
	measurable.TLSHandshaker
	*Saver
}

// Handshake implements measurable.TLSHandshaker.Handshake
func (th TLSHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	th.Saver.emit(Event{Op: "TLSHandshakeStart"})
	tlsconn, state, err := th.TLSHandshaker.Handshake(ctx, conn, config)
	th.Saver.emit(Event{Op: "TLSHandshakeDone"})
	return tlsconn, state, err
}

// HTTPTransport is a loggable HTTPTransport
type HTTPTransport struct {
	measurable.HTTPTransport
	*Saver
}

// RoundTrip implements measurable.HTTPTransport.RoundTrip
func (txp HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	txp.Saver.emit(Event{Op: "RoundTripStart"})
	resp, err := txp.HTTPTransport.RoundTrip(req)
	txp.Saver.emit(Event{Op: "RoundTripDone"})
	return resp, err
}

// WithSaver creates a copy of the provided context that is configured
// to use the specified saver for every dial, request, etc.
func WithSaver(ctx context.Context, saver *Saver) context.Context {
	config := measurable.ContextConfigOrDefault(ctx)
	return measurable.WithConfig(ctx, &measurable.Config{
		Connector: &Connector{
			Connector: config.Connector,
			Saver:     saver,
		},
		HTTPTransport: &HTTPTransport{
			HTTPTransport: config.HTTPTransport,
			Saver:         saver,
		},
		Resolver: &Resolver{
			Resolver: config.Resolver,
			Saver:    saver,
		},
		TLSHandshaker: &TLSHandshaker{
			TLSHandshaker: config.TLSHandshaker,
			Saver:         saver,
		},
	})
}
