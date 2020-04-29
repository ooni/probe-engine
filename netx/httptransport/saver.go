package httptransport

import (
	"net/http"
	"net/http/httptrace"
	"time"

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
