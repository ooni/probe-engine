package resolver

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/netx/trace"
)

// SaverResolver is a resolver that saves events
type SaverResolver struct {
	Resolver
	Saver *trace.Saver
}

// LookupHost implements Resolver.LookupHost
func (r SaverResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	start := time.Now()
	r.Saver.Write(trace.Event{
		Hostname: hostname,
		Name:     "resolve_start",
		Time:     start,
	})
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	stop := time.Now()
	r.Saver.Write(trace.Event{
		Addresses: addrs,
		Duration:  stop.Sub(start),
		Err:       err,
		Hostname:  hostname,
		Name:      "resolve_done",
		Time:      stop,
	})
	return addrs, err
}

// SaverDNSTransport is a DNS transport that saves events
type SaverDNSTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverDNSTransport) RoundTrip(ctx context.Context, query []byte) ([]byte, error) {
	start := time.Now()
	txp.Saver.Write(trace.Event{
		DNSQuery: query,
		Name:     "dns_round_trip_start",
		Time:     start,
	})
	reply, err := txp.RoundTripper.RoundTrip(ctx, query)
	stop := time.Now()
	txp.Saver.Write(trace.Event{
		DNSQuery: query,
		DNSReply: reply,
		Duration: stop.Sub(start),
		Err:      err,
		Name:     "dns_round_trip_done",
		Time:     stop,
	})
	return reply, err
}

var _ Resolver = SaverResolver{}
var _ RoundTripper = SaverDNSTransport{}
