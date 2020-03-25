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

// Saver adds event saving to measurable operations.
type Saver struct {
	measurable.Operations
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

// LookupHost performs an host lookup
func (s *Saver) LookupHost(ctx context.Context, domain string) ([]string, error) {
	s.emit(Event{Op: "LookupHostStart"})
	addrs, err := s.Operations.LookupHost(ctx, domain)
	s.emit(Event{Op: "LookupHostDone"})
	return addrs, err
}

// DialContext establishes a new connection
func (s *Saver) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	s.emit(Event{Op: "DialContextStart"})
	conn, err := s.Operations.DialContext(ctx, network, address)
	s.emit(Event{Op: "DialContextDone"})
	return conn, err
}

// Handshake performs the TLS handshake
func (s *Saver) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	s.emit(Event{Op: "TLSHandshakeStart"})
	tlsconn, state, err := s.Operations.Handshake(ctx, conn, config)
	s.emit(Event{Op: "TLSHandshakeDone"})
	return tlsconn, state, err
}

// RoundTrip performs the HTTP round trip
func (s *Saver) RoundTrip(req *http.Request) (*http.Response, error) {
	s.emit(Event{Op: "RoundTripStart"})
	resp, err := s.Operations.RoundTrip(req)
	s.emit(Event{Op: "RoundTripDone"})
	return resp, err
}
