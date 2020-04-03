package resolver

import (
	"context"
	"errors"
	"net"

	"github.com/miekg/dns"
	"github.com/ooni/probe-engine/atomicx"
)

// RoundTripper represents an abstract DNS transport.
type RoundTripper interface {
	// RoundTrip sends a DNS query and receives the reply.
	RoundTrip(ctx context.Context, query []byte) (reply []byte, err error)

	// RequiresPadding return true for DoH and DoT according to RFC8467
	RequiresPadding() bool

	// Network is the network of the round tripper (e.g. "dot")
	Network() string

	// Address is the address of the round tripper (e.g. "1.1.1.1:853")
	Address() string
}

// OONI is OONI's DNS resolver. It is a simplistic resolver where we
// manually create and sequentially submit queries. It can use all the
// transports for DNS supported by this library, however.
type OONI struct {
	Encoder     Encoder
	Decoder     Decoder
	NumTimeouts *atomicx.Int64
	Txp         RoundTripper
}

// NewOONIResolver creates a new OONI Resolver instance.
func NewOONIResolver(t RoundTripper) OONI {
	return OONI{
		Encoder:     MiekgEncoder{},
		Decoder:     MiekgDecoder{},
		NumTimeouts: atomicx.NewInt64(),
		Txp:         t,
	}
}

// Transport returns the transport being used.
func (r OONI) Transport() RoundTripper {
	return r.Txp
}

// LookupHost returns the IP addresses of a host
func (r OONI) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	var addrs []string
	addrsA, errA := r.roundTripWithRetry(ctx, hostname, dns.TypeA)
	addrsAAAA, errAAAA := r.roundTripWithRetry(ctx, hostname, dns.TypeAAAA)
	if errA != nil && errAAAA != nil {
		return nil, errA
	}
	addrs = append(addrs, addrsA...)
	addrs = append(addrs, addrsAAAA...)
	return addrs, nil
}

func (r OONI) roundTripWithRetry(
	ctx context.Context, hostname string, qtype uint16) ([]string, error) {
	var errorslist []error
	for i := 0; i < 3; i++ {
		replies, err := r.roundTrip(ctx, hostname, qtype)
		if err == nil {
			return replies, nil
		}
		errorslist = append(errorslist, err)
		var operr *net.OpError
		if errors.As(err, &operr) == false || operr.Timeout() == false {
			// The first error is the one that is most likely to be caused
			// by the network. Subsequent errors are more likely to be caused
			// by context deadlines. So, the first error is attached to an
			// operation, while subsequent errors may possibly not be. If
			// so, the resulting failing operation is not correct.
			break
		}
		r.NumTimeouts.Add(1)
	}
	// bugfix: we MUST return one of the errors otherwise we confuse the
	// mechanism in errwrap that classifies the root cause operation, since
	// it would not be able to find a child with a major operation error
	return nil, errorslist[0]
}

func (r OONI) roundTrip(
	ctx context.Context, hostname string, qtype uint16) ([]string, error) {
	querydata, err := r.Encoder.Encode(hostname, qtype, r.Txp.RequiresPadding())
	if err != nil {
		return nil, err
	}
	replydata, err := r.Txp.RoundTrip(ctx, querydata)
	if err != nil {
		return nil, err
	}
	return r.Decoder.Decode(qtype, replydata)
}
