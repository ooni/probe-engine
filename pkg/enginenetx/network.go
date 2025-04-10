package enginenetx

//
// Network - the top-level object of this package, used by the
// OONI engine to communicate with several backends
//

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/pkg/bytecounter"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
	"github.com/ooni/probe-engine/pkg/runtimex"
	"golang.org/x/net/publicsuffix"
)

// Network is the network abstraction used by the OONI engine.
//
// The zero value is invalid; construct using the [NewNetwork] func.
type Network struct {
	reso  model.Resolver
	stats *statsManager
	txp   model.HTTPTransport
}

// HTTPTransport returns the underlying [model.HTTPTransport].
func (n *Network) HTTPTransport() model.HTTPTransport {
	return n.txp
}

// NewHTTPClient is a convenience function for building an [*http.Client] using
// the underlying [model.HTTPTransport] and the correct cookies configuration.
func (n *Network) NewHTTPClient() *http.Client {
	// Note: cookiejar.New cannot fail, so we're using runtimex.Try1 here
	return &http.Client{
		Transport: n.txp,
		Jar: runtimex.Try1(cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		})),
	}
}

// Close ensures that we close idle connections and persist statistics.
func (n *Network) Close() error {
	// TODO(bassosimone): do we want to introduce "once" semantics in this method? It
	// does not seem necessary since there's no resource we can close just once.

	// make sure we close the transport's idle connections
	n.txp.CloseIdleConnections()

	// same as above but for the resolver's connections
	n.reso.CloseIdleConnections()

	// make sure we sync stats to disk and shutdown the background trimmer
	return n.stats.Close()
}

// NewNetwork creates a new [*Network] for the engine. This network MUST NOT be
// used for measuring because it implements engine-specific policies.
//
// You MUST call the Close method when done using the network. This method ensures
// that (i) we close idle connections and (ii) persist statistics.
//
// Arguments:
//
// - counter is the [*bytecounter.Counter] to use;
//
// - kvStore is a [model.KeyValueStore] for persisting stats;
//
// - logger is the [model.Logger] to use;
//
// - proxyURL is the OPTIONAL proxy URL;
//
// - resolver is the [model.Resolver] to use.
//
// The presence of the proxyURL MAY cause this function to possibly build a
// network with different behavior with respect to circumvention. If there is
// an upstream proxy we're going to trust it is doing circumvention for us.
func NewNetwork(
	counter *bytecounter.Counter,
	kvStore model.KeyValueStore,
	logger model.Logger,
	proxyURL *url.URL,
	resolver model.Resolver,
) *Network {
	// Create a dialer ONLY used for dialing unencrypted TCP connections. The common use
	// case of this Network is to dial encrypted connections. For this reason, here it is
	// reasonably fine to use the legacy sequential dialer implemented in netxlite.
	netx := &netxlite.Netx{}
	dialer := netx.NewDialerWithResolver(logger, resolver)

	// Create manager for keeping track of statistics. This implies creating a background
	// goroutine that we'll need to close when we're done.
	const trimInterval = 30 * time.Second
	stats := newStatsManager(kvStore, logger, trimInterval)

	// Create a TLS dialer ONLY used for dialing TLS connections. This dialer will use
	// happy-eyeballs and possibly custom policies for dialing TLS connections.
	httpsDialer := newHTTPSDialer(
		logger,
		&netxlite.Netx{Underlying: nil}, // nil means using netxlite's singleton
		newHTTPSDialerPolicy(kvStore, logger, proxyURL, resolver, stats),
		stats,
	)

	// Here we're creating a "new style" HTTPS transport, which has less
	// restrictions compared to the "old style" one.
	//
	// Note that:
	//
	// - we're enabling compression, which is desiredable since this transport
	// is not made for measuring and compression is good(TM), also note that
	// when the request uses Accept-Encoding, this kind of automatic management
	// of compression is disabled, so there is no conflict.
	//
	// - if proxyURL is nil, the proxy option is equivalent to disabling
	// the proxy, otherwise it means that we're using the net/http library
	// to dial for proxies, which has some restrictions.
	//
	// - this code does not work as intended when using netem and proxies
	// as documented by TODO(https://github.com/ooni/probe/issues/2536).
	txp := netxlite.NewHTTPTransportWithOptions(
		logger, dialer, httpsDialer,
		netxlite.HTTPTransportOptionDisableCompression(false),
		netxlite.HTTPTransportOptionProxyURL(proxyURL),
	)

	// Make sure we count the bytes sent and received as part of the session
	txp = bytecounter.WrapHTTPTransport(txp, counter)

	network := &Network{
		reso:  resolver,
		stats: stats,
		txp:   txp,
	}
	return network
}

// newHTTPSDialerPolicy contains the logic to select the [HTTPSDialerPolicy] to use.
func newHTTPSDialerPolicy(
	kvStore model.KeyValueStore,
	logger model.Logger,
	proxyURL *url.URL, // optional!
	resolver model.Resolver,
	stats *statsManager,
) httpsDialerPolicy {
	// in case there's a proxy URL, we're going to trust the proxy to do the right thing and
	// know what it's doing, hence we'll have a very simple DNS policy
	if proxyURL != nil {
		return &dnsPolicy{logger, resolver}
	}

	// create a policy interleaving stats policies and bridges policies
	statsOrBridges := &mixPolicyInterleave{
		Primary: &statsPolicyV2{
			Stats: stats,
		},
		Fallback: &bridgesPolicyV2{},
		Factor:   3,
	}

	// wrap the DNS policy with a policy that extends tactics for test
	// helpers so that we also try using different SNIs.
	dnsExt := &testHelpersPolicy{
		Child: &dnsPolicy{logger, resolver},
	}

	// compose dnsExt and statsOrBridges such that dnsExt has
	// priority in the selection of tactics
	composed := &mixPolicyInterleave{
		Primary:  dnsExt,
		Fallback: statsOrBridges,
		Factor:   3,
	}

	// attempt to load a user-provided dialing policy
	primary, err := newUserPolicyV2(kvStore)

	// on error, just use composed
	if err != nil {
		return composed
	}

	// otherwise, finish creating the dialing policy
	policy := &mixPolicyEitherOr{
		Primary:  primary,
		Fallback: composed,
	}

	return policy
}
