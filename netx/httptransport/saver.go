package httptransport

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/ooni/probe-engine/netx/trace"
)

// SaverPerformanceHTTPTransport is a RoundTripper that saves
// performance events occurring during the round trip
type SaverPerformanceHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverPerformanceHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tracep := httptrace.ContextClientTrace(req.Context())
	if tracep == nil {
		tracep = &httptrace.ClientTrace{
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
	}
	return txp.RoundTripper.RoundTrip(req)
}

// SaverRoundTripHTTPTransport is a RoundTripper that saves base
// events pertaining to the HTTP round trip
type SaverRoundTripHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverRoundTripHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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

// SaverBodyHTTPTransport is a RoundTripper that saves
// body events occurring during the round trip
type SaverBodyHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverBodyHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	const snapsize = 1 << 17
	if req.Body != nil {
		txp.Saver.Write(trace.Event{
			HTTPRequestBody: saverReadSnap(&req.Body, snapsize),
			Name:            "http_request_body_snapshot",
			Time:            time.Now(),
		})
	}
	resp, err := txp.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	txp.Saver.Write(trace.Event{
		HTTPResponseBody: saverReadSnap(&resp.Body, snapsize),
		Name:             "http_response_body_snapshot",
		Time:             time.Now(),
	})
	return resp, nil
}

func saverReadSnap(r *io.ReadCloser, snapsize int64) *trace.Snapshot {
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

var _ RoundTripper = SaverPerformanceHTTPTransport{}
var _ RoundTripper = SaverRoundTripHTTPTransport{}
var _ RoundTripper = SaverBodyHTTPTransport{}
