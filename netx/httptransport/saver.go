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

// SaverHTTPTransport is a RoundTripper that saves events
type SaverHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
	start := time.Now()
	txp.Saver.Write(trace.Event{
		HTTPRequestBody: reqbody,
		HTTPRequest:     req,
		Name:            "http_round_trip_start",
		Time:            start,
	})
	resp, err := txp.RoundTripper.RoundTrip(req)
	stop := time.Now()
	respbody := saverReadResponseSnap(resp, snapsize)
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
