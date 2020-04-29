package httptransport

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
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
	const snapsize = 1 << 17
	reqbody := saverReadSnap(&req.Body, snapsize)
	start := time.Now() // exclude time to read body snapshot
	txp.Saver.Write(trace.Event{
		HTTPRequestBody: reqbody,
		HTTPRequest:     req,
		Name:            "http_round_trip_start",
		Time:            start,
	})
	resp, err := txp.RoundTripper.RoundTrip(req)
	respbody := saverReadResponseSnap(resp, snapsize)
	stop := time.Now() // include time to read body snapshot
	txp.Saver.Write(trace.Event{
		Duration:         stop.Sub(start),
		Err:              err,
		HTTPRequestBody:  reqbody,
		HTTPRequest:      req,
		HTTPResponseBody: respbody,
		HTTPResponse:     resp,
		Name:             "http_round_trip_done",
		Time:             stop,
	})
	return resp, err
}

func saverReadResponseSnap(resp *http.Response, snapsize int64) *trace.Snapshot {
	if resp == nil {
		return nil
	}
	return saverReadSnap(&resp.Body, snapsize)
}

func saverReadSnap(r *io.ReadCloser, snapsize int64) *trace.Snapshot {
	if r == nil || *r == nil {
		return nil
	}
	data, err := ioutil.ReadAll(io.LimitReader(*r, snapsize))
	if err != nil {
		*r = saverReadCloser{
			Closer: *r,
			Reader: io.MultiReader(bytes.NewReader(data), saverErrReader{err}),
		}
		return nil
	}
	*r = saverReadCloser{
		Closer: *r,
		Reader: io.MultiReader(bytes.NewReader(data), *r),
	}
	return &trace.Snapshot{Data: data, Limit: snapsize}
}

type saverReadCloser struct {
	io.Closer
	io.Reader
}

type saverErrReader struct {
	Err error
}

func (r saverErrReader) Read(p []byte) (int, error) {
	return 0, r.Err
}

var _ RoundTripper = SaverHTTPTransport{}
