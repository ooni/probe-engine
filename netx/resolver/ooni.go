package resolver

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/miekg/dns"
	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// OONIResolver is OONI's DNS client. It is a simplistic client where we
// manually create and submit queries. It can use all the transports
// for DNS supported by this library, however.
type OONIResolver struct {
	ntimeouts *atomicx.Int64
	transport modelx.DNSRoundTripper
}

// NewOONIResolver creates a new OONI Resolver instance.
func NewOONIResolver(t modelx.DNSRoundTripper) *OONIResolver {
	return &OONIResolver{
		ntimeouts: atomicx.NewInt64(),
		transport: t,
	}
}

// Transport returns the transport being used.
func (c *OONIResolver) Transport() modelx.DNSRoundTripper {
	return c.transport
}

var errNotImpl = errors.New("Not implemented")

// LookupHost returns the IP addresses of a host
func (c *OONIResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	var addrs []string
	var reply *dns.Msg
	reply, errA := c.roundTripWithRetry(ctx, hostname, dns.TypeA)
	if errA == nil {
		for _, answer := range reply.Answer {
			if rra, ok := answer.(*dns.A); ok {
				ip := rra.A
				addrs = append(addrs, ip.String())
			}
		}
	}
	reply, errAAAA := c.roundTripWithRetry(ctx, hostname, dns.TypeAAAA)
	if errAAAA == nil {
		for _, answer := range reply.Answer {
			if rra, ok := answer.(*dns.AAAA); ok {
				ip := rra.AAAA
				addrs = append(addrs, ip.String())
			}
		}
	}
	return lookupHostResult(addrs, errA, errAAAA)
}

func lookupHostResult(addrs []string, errA, errAAAA error) ([]string, error) {
	if len(addrs) > 0 {
		return addrs, nil
	}
	if errA != nil {
		return nil, errA
	}
	if errAAAA != nil {
		return nil, errAAAA
	}
	return nil, errors.New("ooniresolver: no response returned")
}

const (
	// desiredBlockSize is the size that the padded query should be multiple of
	desiredBlockSize = 128

	// maxResponseSize is the maximum response size for EDNS0
	maxResponseSize = 4096

	// dnssecEnabled turns on support for DNSSEC when using EDNS0
	dnssecEnabled = true
)

func (c *OONIResolver) newQueryWithQuestion(q dns.Question, needspadding bool) (query *dns.Msg) {
	query = new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = make([]dns.Question, 1)
	query.Question[0] = q
	if needspadding {
		query.SetEdns0(maxResponseSize, dnssecEnabled)
		// Clients SHOULD pad queries to the closest multiple of
		// 128 octets RFC8467#section-4.1. We inflate the query
		// length by the size of the option (i.e. 4 octets). The
		// cast to uint is necessary to make the modulus operation
		// work as intended when the desiredBlockSize is smaller
		// than (query.Len()+4) ¯\_(ツ)_/¯.
		remainder := (desiredBlockSize - uint(query.Len()+4)) % desiredBlockSize
		opt := new(dns.EDNS0_PADDING)
		opt.Padding = make([]byte, remainder)
		query.IsEdns0().Option = append(query.IsEdns0().Option, opt)
	}
	return
}

func (c *OONIResolver) roundTripWithRetry(
	ctx context.Context, hostname string, qtype uint16,
) (*dns.Msg, error) {
	var errorslist []error
	for i := 0; i < 3; i++ {
		reply, err := c.roundTrip(ctx, c.newQueryWithQuestion(dns.Question{
			Name:   dns.Fqdn(hostname),
			Qtype:  qtype,
			Qclass: dns.ClassINET,
		}, c.Transport().RequiresPadding()))
		if err == nil {
			return reply, nil
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
		c.ntimeouts.Add(1)
	}
	// bugfix: we MUST return one of the errors otherwise we confuse the
	// mechanism in errwrap that classifies the root cause operation, since
	// it would not be able to find a child with a major operation error
	return nil, errorslist[0]
}

func (c *OONIResolver) roundTrip(ctx context.Context, query *dns.Msg) (reply *dns.Msg, err error) {
	return c.mockableRoundTrip(
		ctx, query, func(msg *dns.Msg) ([]byte, error) {
			return msg.Pack()
		},
		func(t modelx.DNSRoundTripper, query []byte) (reply []byte, err error) {
			// Pass ctx to round tripper for cancellation as well
			// as to propagate context information
			return t.RoundTrip(ctx, query)
		},
		func(msg *dns.Msg, data []byte) (err error) {
			return msg.Unpack(data)
		},
	)
}

func (c *OONIResolver) mockableRoundTrip(
	ctx context.Context,
	query *dns.Msg,
	pack func(msg *dns.Msg) ([]byte, error),
	roundTrip func(t modelx.DNSRoundTripper, query []byte) (reply []byte, err error),
	unpack func(msg *dns.Msg, data []byte) (err error),
) (reply *dns.Msg, err error) {
	var (
		querydata []byte
		replydata []byte
	)
	querydata, err = pack(query)
	if err != nil {
		return
	}
	root := modelx.ContextMeasurementRootOrDefault(ctx)
	root.Handler.OnMeasurement(modelx.Measurement{
		DNSQuery: &modelx.DNSQueryEvent{
			Data:                   querydata,
			DialID:                 dialid.ContextDialID(ctx),
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
			Msg:                    query,
		},
	})
	replydata, err = roundTrip(c.transport, querydata)
	if err != nil {
		return
	}
	reply = new(dns.Msg)
	err = unpack(reply, replydata)
	if err != nil {
		return
	}
	root.Handler.OnMeasurement(modelx.Measurement{
		DNSReply: &modelx.DNSReplyEvent{
			Data:                   replydata,
			DialID:                 dialid.ContextDialID(ctx),
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
			Msg:                    reply,
		},
	})
	err = mapError(reply.Rcode)
	return
}

func mapError(rcode int) error {
	// TODO(bassosimone): map more errors to net.DNSError names
	switch rcode {
	case dns.RcodeSuccess:
		return nil
	case dns.RcodeNameError:
		return errors.New("ooniresolver: no such host")
	default:
		return errors.New("ooniresolver: query failed")
	}
}
