// Package bodytracer contains the HTTP body tracer. The purpose
// of tracing is to emit events while we read response bodies.
package bodytracer

import (
	"io"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// Transport performs single HTTP transactions and emits
// measurement events as they happen.
type Transport struct {
	roundTripper http.RoundTripper
}

// New creates a new Transport.
func New(roundTripper http.RoundTripper) *Transport {
	return &Transport{roundTripper: roundTripper}
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.roundTripper.RoundTrip(req)
	if err != nil {
		return
	}
	// "The http Client and Transport guarantee that Body is always
	//  non-nil, even on responses without a body or responses with
	//  a zero-length body." (from the docs)
	resp.Body = &bodyWrapper{
		ReadCloser: resp.Body,
		root:       modelx.ContextMeasurementRootOrDefault(req.Context()),
		tid:        transactionid.ContextTransactionID(req.Context()),
	}
	return
}

// CloseIdleConnections closes the idle connections.
func (t *Transport) CloseIdleConnections() {
	// Adapted from net/http code
	type closeIdler interface {
		CloseIdleConnections()
	}
	if tr, ok := t.roundTripper.(closeIdler); ok {
		tr.CloseIdleConnections()
	}
}

type bodyWrapper struct {
	io.ReadCloser
	root *modelx.MeasurementRoot
	tid  int64
}

func (bw *bodyWrapper) Read(b []byte) (n int, err error) {
	n, err = bw.ReadCloser.Read(b)
	bw.root.Handler.OnMeasurement(modelx.Measurement{
		HTTPResponseBodyPart: &modelx.HTTPResponseBodyPartEvent{
			// "Read reads up to len(p) bytes into p. It returns the number of
			// bytes read (0 <= n <= len(p)) and any error encountered."
			Data:                   b[:n],
			Error:                  err,
			DurationSinceBeginning: time.Now().Sub(bw.root.Beginning),
			TransactionID:          bw.tid,
		},
	})
	return
}

func (bw *bodyWrapper) Close() (err error) {
	err = bw.ReadCloser.Close()
	bw.root.Handler.OnMeasurement(modelx.Measurement{
		HTTPResponseDone: &modelx.HTTPResponseDoneEvent{
			DurationSinceBeginning: time.Now().Sub(bw.root.Beginning),
			TransactionID:          bw.tid,
		},
	})
	return
}
