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

// SaverMetadataHTTPTransport is a RoundTripper that saves
// events related to HTTP request and response metadata
type SaverMetadataHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverMetadataHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	txp.Saver.Write(trace.Event{
		HTTPHeaders: req.Header,
		HTTPMethod:  req.Method,
		HTTPURL:     req.URL.String(),
		Name:        "http_request_metadata",
		Time:        time.Now(),
	})
	resp, err := txp.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	txp.Saver.Write(trace.Event{
		HTTPHeaders:    resp.Header,
		HTTPStatusCode: resp.StatusCode,
		Name:           "http_response_metadata",
		Time:           time.Now(),
	})
	return resp, err
}

// SaverTransactionHTTPTransport is a RoundTripper that saves
// events related to the HTTP transaction
type SaverTransactionHTTPTransport struct {
	RoundTripper
	Saver *trace.Saver
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverTransactionHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	txp.Saver.Write(trace.Event{
		Name: "http_transaction_start",
		Time: time.Now(),
	})
	resp, err := txp.RoundTripper.RoundTrip(req)
	txp.Saver.Write(trace.Event{
		Err:  err,
		Name: "http_transaction_done",
		Time: time.Now(),
	})
	return resp, err
}

// SaverBodyHTTPTransport is a RoundTripper that saves
// body events occurring during the round trip
type SaverBodyHTTPTransport struct {
	RoundTripper
	Saver        *trace.Saver
	SnapshotSize int
}

// RoundTrip implements RoundTripper.RoundTrip
func (txp SaverBodyHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	const defaultSnapSize = 1 << 17
	snapsize := defaultSnapSize
	if txp.SnapshotSize != 0 {
		snapsize = txp.SnapshotSize
	}
	if req.Body != nil {
		data, err := saverSnapRead(req.Body, snapsize)
		if err != nil {
			return nil, err
		}
		req.Body = saverCompose(data, req.Body)
		txp.Saver.Write(trace.Event{
			DataIsTruncated: len(data) >= snapsize,
			Data:            data,
			Name:            "http_request_body_snapshot",
			Time:            time.Now(),
		})
	}
	resp, err := txp.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	data, err := saverSnapRead(resp.Body, snapsize)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body = saverCompose(data, resp.Body)
	txp.Saver.Write(trace.Event{
		DataIsTruncated: len(data) >= snapsize,
		Data:            data,
		Name:            "http_response_body_snapshot",
		Time:            time.Now(),
	})
	return resp, nil
}

func saverSnapRead(r io.ReadCloser, snapsize int) ([]byte, error) {
	return ioutil.ReadAll(io.LimitReader(r, int64(snapsize)))
}

func saverCompose(data []byte, r io.ReadCloser) io.ReadCloser {
	return saverReadCloser{Closer: r, Reader: io.MultiReader(bytes.NewReader(data), r)}
}

type saverReadCloser struct {
	io.Closer
	io.Reader
}

var _ RoundTripper = SaverPerformanceHTTPTransport{}
var _ RoundTripper = SaverMetadataHTTPTransport{}
var _ RoundTripper = SaverBodyHTTPTransport{}
var _ RoundTripper = SaverTransactionHTTPTransport{}
