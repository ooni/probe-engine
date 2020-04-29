package httptransport

import (
	"net/http"
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

var _ RoundTripper = SaverHTTPTransport{}
