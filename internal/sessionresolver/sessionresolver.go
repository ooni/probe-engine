// Package sessionresolver contains the resolver used by the session. This
// resolver uses Powerdns DoH by default and falls back on the system
// provided resolver if Powerdns DoH is not working.
package sessionresolver

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/netx/httptransport"
)

// Resolver is the session resolver.
type Resolver struct {
	Primary         httptransport.DNSClient
	PrimaryFailure  *atomicx.Int64
	Fallback        httptransport.DNSClient
	FallbackFailure *atomicx.Int64
}

// New creates a new session resolver.
func New(config httptransport.Config) *Resolver {
	primary, err := httptransport.NewDNSClient(config, "doh://powerdns")
	runtimex.PanicOnError(err, "cannot create powerdns resolver")
	fallback, err := httptransport.NewDNSClient(config, "system:///")
	runtimex.PanicOnError(err, "cannot create system resolver")
	return &Resolver{
		Primary:         primary,
		PrimaryFailure:  atomicx.NewInt64(),
		Fallback:        fallback,
		FallbackFailure: atomicx.NewInt64(),
	}
}

// CloseIdleConnections closes the idle connections, if any
func (r *Resolver) CloseIdleConnections() {
	r.Primary.CloseIdleConnections()
	r.Fallback.CloseIdleConnections()
}

// LookupHost implements Resolver.LookupHost
func (r *Resolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	// We doing this like Firefox trr 2 mode. See here:
	// https://wiki.mozilla.org/Trusted_Recursive_Resolver#DNS-over-HTTPS_Prefs_in_Firefox
	// And here: network.trr.request_timeout_ms
	// But we use a higher timeout than Firefox to be safe and use DoH more
	// often if possible
	trr2, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	addrs, err := r.Primary.LookupHost(trr2, hostname)
	if err != nil {
		r.PrimaryFailure.Add(1)
		addrs, err = r.Fallback.LookupHost(ctx, hostname)
		if err != nil {
			r.FallbackFailure.Add(1)
		}
	}
	return addrs, err
}

// Network implements Resolver.Network
func (r *Resolver) Network() string {
	return "sessionresolver"
}

// Address implements Resolver.Address
func (r *Resolver) Address() string {
	return ""
}
