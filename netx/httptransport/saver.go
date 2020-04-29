package httptransport

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/trace"
)

// SaverPerformanceHTTPTransport is a RoundTripper that saves
// performance events during the round trip
type SaverPerformanceHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverPerformanceHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Implementation description: create a new context independent from
	// the original one but still honour the original one. This guarantees
	// that we will always have a single tracer.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	origCtx := req.Context()
	// Make sure we honour the parent's byte counters. This is horrible
	// and if we could avoid using context for doing that...
	if bc := dialer.ContextSessionByteCounter(origCtx); bc != nil {
		ctx = dialer.WithSessionByteCounter(ctx, bc)
	}
	if bc := dialer.ContextExperimentByteCounter(origCtx); bc != nil {
		ctx = dialer.WithExperimentByteCounter(ctx, bc)
	}
	req = req.WithContext(ctx)
	respch := make(chan *http.Response)
	errch := make(chan error)
	go txp.roundTrip(req, respch, errch)
	select {
	case <-origCtx.Done():
		return nil, origCtx.Err()
	case resp := <-respch:
		return resp, nil
	case err := <-errch:
		return nil, err
	}
}

func (txp SaverPerformanceHTTPTransport) roundTrip(
	req *http.Request, respch chan *http.Response, errch chan error) {
	// Implementation description: if we cannot write, it means
	// that the parent has already cancelled the context
	resp, err := txp.doRoundTrip(req)
	if err != nil {
		select {
		case errch <- err:
		default:
		}
		return
	}
	select {
	case respch <- resp:
	default:
		resp.Body.Close() // honour round tripper protocol
	}
}

func (txp SaverPerformanceHTTPTransport) doRoundTrip(req *http.Request) (*http.Response, error) {
	// Unconditionally attach tracer. We know there's only this one because
	// above we created a pristine context for this request.
	tracep := &httptrace.ClientTrace{
		WroteHeaders: func() {
			txp.Saver.Write(trace.Event{Name: "http_wrote_headers", Time: time.Now()})
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			txp.Saver.Write(trace.Event{Name: "http_wrote_request", Time: time.Now()})
		},
		GotFirstResponseByte: func() {
			txp.Saver.Write(trace.Event{
				Name: "http_first_response_byte", Time: time.Now()})
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), tracep))
	return txp.RoundTripper.RoundTrip(req)
}

// SaverHTTPTransport is a RoundTripper that saves events
type SaverHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	txp.Saver.Write(trace.Event{
		HTTPRequest: req,
		Name:        "http_round_trip_start",
		Time:        start,
	})
	resp, err := txp.RoundTripper.RoundTrip(req)
	stop := time.Now()
	txp.Saver.Write(trace.Event{
		Duration:     stop.Sub(start),
		Err:          err,
		HTTPRequest:  req,
		HTTPResponse: resp,
		Name:         "http_round_trip_done",
		Time:         stop,
	})
	return resp, err
}

var _ RoundTripper = SaverPerformanceHTTPTransport{}
var _ RoundTripper = SaverHTTPTransport{}
