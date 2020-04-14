package httptransport

import (
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/ooni/probe-engine/netx/trace"
)

// SaverHTTPTransport is a RoundTripper that saves events
type SaverHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
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
	}))
	start := time.Now()
	resp, err := txp.RoundTripper.RoundTrip(req)
	stop := time.Now()
	txp.Saver.Write(trace.Event{
		Duration:     stop.Sub(start),
		Err:          err,
		Name:         "http_round_trip",
		HTTPRequest:  req,
		HTTPResponse: resp,
		Time:         stop,
	})
	return resp, err
}

var _ RoundTripper = SaverHTTPTransport{}
