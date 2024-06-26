package enginenetx

//
// HTTPS dialing policy where we generate tactics in the usual way
// by using a DNS resolver and using SNI == VerifyHostname
//

import (
	"context"

	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

// dnsPolicy is the default TLS dialing policy where we use the
// given resolver and the domain as the SNI.
//
// The zero value is invalid; please, init all MANDATORY fields.
type dnsPolicy struct {
	// Logger is the MANDATORY logger.
	Logger model.Logger

	// Resolver is the MANDATORY resolver.
	Resolver model.Resolver
}

var _ httpsDialerPolicy = &dnsPolicy{}

// LookupTactics implements httpsDialerPolicy.
func (p *dnsPolicy) LookupTactics(
	ctx context.Context, domain, port string) <-chan *httpsDialerTactic {
	out := make(chan *httpsDialerTactic)

	go func() {
		// make sure we close the output channel when done
		// so the reader knows that we're done
		defer close(out)

		// Do not even start the DNS lookup if the context has already been canceled, which
		// happens if some policy running before us had successfully connected
		if err := ctx.Err(); err != nil {
			p.Logger.Debugf("dnsPolicy: LookupTactics: %s", err.Error())
			return
		}

		// See https://github.com/ooni/probe-cli/pull/1295#issuecomment-1731243994 for context
		// on why here we MUST make sure we short-circuit IP addresses.
		resoWithShortCircuit := &netxlite.ResolverShortCircuitIPAddr{Resolver: p.Resolver}

		addrs, err := resoWithShortCircuit.LookupHost(ctx, domain)
		if err != nil {
			p.Logger.Warnf("resoWithShortCircuit.LookupHost: %s", err.Error())
			return
		}

		// The tactics we generate here have SNI == VerifyHostname == domain
		for _, addr := range addrs {
			tactic := &httpsDialerTactic{
				Address:        addr,
				InitialDelay:   0, // set when dialing
				Port:           port,
				SNI:            domain,
				VerifyHostname: domain,
			}
			out <- tactic
		}
	}()

	return out
}
