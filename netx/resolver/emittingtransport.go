package resolver

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// EmittingTransport emits round trip events
type EmittingTransport struct {
	RoundTripper
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp EmittingTransport) RoundTrip(ctx context.Context, querydata []byte) ([]byte, error) {
	root := modelx.ContextMeasurementRootOrDefault(ctx)
	root.Handler.OnMeasurement(modelx.Measurement{
		DNSQuery: &modelx.DNSQueryEvent{
			Data:                   querydata,
			DialID:                 dialid.ContextDialID(ctx),
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
		},
	})
	replydata, err := txp.RoundTripper.RoundTrip(ctx, querydata)
	if err != nil {
		return nil, err
	}
	root.Handler.OnMeasurement(modelx.Measurement{
		DNSReply: &modelx.DNSReplyEvent{
			Data:                   replydata,
			DialID:                 dialid.ContextDialID(ctx),
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
		},
	})
	return replydata, nil
}
