package resolver

import (
	"context"
	"errors"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// ParentResolver is the emitter resolver
type ParentResolver struct {
	bogonsCount *atomicx.Int64
	resolver    Resolver
}

// NewParentResolver creates a new emitter resolver
func NewParentResolver(resolver Resolver) *ParentResolver {
	return &ParentResolver{
		bogonsCount: atomicx.NewInt64(),
		resolver:    resolver,
	}
}

type queryableTransport interface {
	Network() string
	Address() string
	RequiresPadding() bool
}

type queryableResolver interface {
	Transport() RoundTripper
}

func (r *ParentResolver) queryTransport() (network string, address string) {
	if reso, okay := r.resolver.(queryableResolver); okay {
		if transport, okay := reso.Transport().(queryableTransport); okay {
			network, address = transport.Network(), transport.Address()
		}
	}
	return
}

// LookupHost returns the IP addresses of a host
func (r *ParentResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	network, address := r.queryTransport()
	dialID := dialid.ContextDialID(ctx)
	txID := transactionid.ContextTransactionID(ctx)
	root := modelx.ContextMeasurementRootOrDefault(ctx)
	root.Handler.OnMeasurement(modelx.Measurement{
		ResolveStart: &modelx.ResolveStartEvent{
			DialID:                 dialID,
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
			Hostname:               hostname,
			TransactionID:          txID,
			TransportAddress:       address,
			TransportNetwork:       network,
		},
	})
	addrs, err := r.lookupHost(ctx, hostname)
	containsBogons := errors.Is(err, modelx.ErrDNSBogon)
	if containsBogons {
		// By default root.ErrDNSBogon is nil. Treating bogons as
		// errors could prevent us from measuring, e.g., legitimate
		// internal-only servers in Iran. This is why we have not
		// enabled this functionality by default. Of course, it is
		// instead smart to treat bogons as errors when we're using
		// a website that we _know_ cannot have bogons.
		//
		// See also <https://github.com/ooni/probe-engine/netx/issues/126>.
		err = root.ErrDNSBogon
	}
	err = errwrapper.SafeErrWrapperBuilder{
		DialID:        dialID,
		Error:         err,
		Operation:     "resolve",
		TransactionID: txID,
	}.MaybeBuild()
	root.Handler.OnMeasurement(modelx.Measurement{
		ResolveDone: &modelx.ResolveDoneEvent{
			Addresses:              addrs,
			ContainsBogons:         containsBogons,
			DialID:                 dialID,
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
			Error:                  err,
			Hostname:               hostname,
			TransactionID:          txID,
			TransportAddress:       address,
			TransportNetwork:       network,
		},
	})
	// Respect general Go expectation that one doesn't return
	// both a value and a non-nil error
	if errors.Is(err, modelx.ErrDNSBogon) {
		addrs = nil
	}
	return addrs, err
}

func (r *ParentResolver) lookupHost(ctx context.Context, hostname string) ([]string, error) {
	addrs, err := r.resolver.LookupHost(ctx, hostname)
	for _, addr := range addrs {
		if IsBogon(addr) == true {
			return r.detectedBogon(ctx, hostname, addrs)
		}
	}
	return addrs, err
}

func (r *ParentResolver) detectedBogon(
	ctx context.Context, hostname string, addrs []string,
) ([]string, error) {
	r.bogonsCount.Add(1)
	return addrs, modelx.ErrDNSBogon
}
