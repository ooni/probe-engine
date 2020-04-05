package oldhttptransport

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/netx/internal/connid"
	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// TraceTripper performs single HTTP transactions.
type TraceTripper struct {
	readAllErrs  *atomicx.Int64
	readAll      func(r io.Reader) ([]byte, error)
	roundTripper http.RoundTripper
}

// NewTraceTripper creates a new Transport.
func NewTraceTripper(roundTripper http.RoundTripper) *TraceTripper {
	return &TraceTripper{
		readAllErrs:  atomicx.NewInt64(),
		readAll:      ioutil.ReadAll,
		roundTripper: roundTripper,
	}
}

type readCloseWrapper struct {
	closer io.Closer
	reader io.Reader
}

func newReadCloseWrapper(
	reader io.Reader, closer io.ReadCloser,
) *readCloseWrapper {
	return &readCloseWrapper{
		closer: closer,
		reader: reader,
	}
}

func (c *readCloseWrapper) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *readCloseWrapper) Close() error {
	return c.closer.Close()
}

func readSnap(
	source *io.ReadCloser, limit int64,
	readAll func(r io.Reader) ([]byte, error),
) (data []byte, err error) {
	data, err = readAll(io.LimitReader(*source, limit))
	if err == nil {
		*source = newReadCloseWrapper(
			io.MultiReader(bytes.NewReader(data), *source),
			*source,
		)
	}
	return
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (t *TraceTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	root := modelx.ContextMeasurementRootOrDefault(req.Context())

	tid := transactionid.ContextTransactionID(req.Context())
	root.Handler.OnMeasurement(modelx.Measurement{
		HTTPRoundTripStart: &modelx.HTTPRoundTripStartEvent{
			DialID:                 dialid.ContextDialID(req.Context()),
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
			Method:                 req.Method,
			TransactionID:          tid,
			URL:                    req.URL.String(),
		},
	})

	var (
		err              error
		majorOp          = "http_round_trip"
		majorOpMu        sync.Mutex
		requestBody      []byte
		requestHeaders   = http.Header{}
		requestHeadersMu sync.Mutex
		snapSize         = modelx.ComputeBodySnapSize(root.MaxBodySnapSize)
	)

	// Save a snapshot of the request body
	if req.Body != nil {
		requestBody, err = readSnap(&req.Body, snapSize, t.readAll)
		if err != nil {
			return nil, err
		}
	}

	// Prepare a tracer for delivering events
	tracer := &httptrace.ClientTrace{
		TLSHandshakeStart: func() {
			majorOpMu.Lock()
			majorOp = "tls_handshake"
			majorOpMu.Unlock()
			// Event emitted by net/http when DialTLS is not
			// configured in the http.Transport
			root.Handler.OnMeasurement(modelx.Measurement{
				TLSHandshakeStart: &modelx.TLSHandshakeStartEvent{
					DurationSinceBeginning: time.Now().Sub(root.Beginning),
					TransactionID:          tid,
				},
			})
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			// Wrapping the error even if we're not returning it because it may
			// less confusing to users to see the wrapped name
			err = errwrapper.SafeErrWrapperBuilder{
				Error:         err,
				Operation:     "tls_handshake",
				TransactionID: tid,
			}.MaybeBuild()
			durationSinceBeginning := time.Now().Sub(root.Beginning)
			// Event emitted by net/http when DialTLS is not
			// configured in the http.Transport
			root.Handler.OnMeasurement(modelx.Measurement{
				TLSHandshakeDone: &modelx.TLSHandshakeDoneEvent{
					ConnectionState:        modelx.NewTLSConnectionState(state),
					Error:                  err,
					DurationSinceBeginning: durationSinceBeginning,
					TransactionID:          tid,
				},
			})
		},
		GotConn: func(info httptrace.GotConnInfo) {
			majorOpMu.Lock()
			majorOp = "http_round_trip"
			majorOpMu.Unlock()
			root.Handler.OnMeasurement(modelx.Measurement{
				HTTPConnectionReady: &modelx.HTTPConnectionReadyEvent{
					ConnID: connid.Compute(
						info.Conn.LocalAddr().Network(),
						info.Conn.LocalAddr().String(),
					),
					DurationSinceBeginning: time.Now().Sub(root.Beginning),
					TransactionID:          tid,
				},
			})
		},
		WroteHeaderField: func(key string, values []string) {
			requestHeadersMu.Lock()
			// Important: do not set directly into the headers map using
			// the [] operator because net/http expects to be able to
			// perform normalization of header names!
			for _, value := range values {
				requestHeaders.Add(key, value)
			}
			requestHeadersMu.Unlock()
			root.Handler.OnMeasurement(modelx.Measurement{
				HTTPRequestHeader: &modelx.HTTPRequestHeaderEvent{
					DurationSinceBeginning: time.Now().Sub(root.Beginning),
					Key:                    key,
					TransactionID:          tid,
					Value:                  values,
				},
			})
		},
		WroteHeaders: func() {
			root.Handler.OnMeasurement(modelx.Measurement{
				HTTPRequestHeadersDone: &modelx.HTTPRequestHeadersDoneEvent{
					DurationSinceBeginning: time.Now().Sub(root.Beginning),
					Headers:                requestHeaders, // [*]
					Method:                 req.Method,     // [*]
					TransactionID:          tid,
					URL:                    req.URL, // [*]
				},
			})
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			// Wrapping the error even if we're not returning it because it may
			// less confusing to users to see the wrapped name
			err := errwrapper.SafeErrWrapperBuilder{
				Error:         info.Err,
				Operation:     "http_round_trip",
				TransactionID: tid,
			}.MaybeBuild()
			root.Handler.OnMeasurement(modelx.Measurement{
				HTTPRequestDone: &modelx.HTTPRequestDoneEvent{
					DurationSinceBeginning: time.Now().Sub(root.Beginning),
					Error:                  err,
					TransactionID:          tid,
				},
			})
		},
		GotFirstResponseByte: func() {
			root.Handler.OnMeasurement(modelx.Measurement{
				HTTPResponseStart: &modelx.HTTPResponseStartEvent{
					DurationSinceBeginning: time.Now().Sub(root.Beginning),
					TransactionID:          tid,
				},
			})
		},
	}

	// If we don't have already a tracer this is a toplevel request, so just
	// set the tracer. Otherwise, we're doing DoH. We cannot set anothert trace
	// because they'd be merged. Instead, replace the existing trace content
	// with the new trace and then remember to reset it.
	origtracer := httptrace.ContextClientTrace(req.Context())
	if origtracer != nil {
		bkp := *origtracer
		*origtracer = *tracer
		defer func() {
			*origtracer = bkp
		}()
	} else {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), tracer))
	}

	resp, err := t.roundTripper.RoundTrip(req)
	err = errwrapper.SafeErrWrapperBuilder{
		Error:         err,
		Operation:     majorOp,
		TransactionID: tid,
	}.MaybeBuild()
	// [*] Require less event joining work by providing info that
	// makes this event alone actionable for OONI
	event := &modelx.HTTPRoundTripDoneEvent{
		DurationSinceBeginning: time.Now().Sub(root.Beginning),
		Error:                  err,
		RequestBodySnap:        requestBody,
		RequestHeaders:         requestHeaders,   // [*]
		RequestMethod:          req.Method,       // [*]
		RequestURL:             req.URL.String(), // [*]
		MaxBodySnapSize:        snapSize,
		TransactionID:          tid,
	}
	if resp != nil {
		event.ResponseHeaders = resp.Header
		event.ResponseStatusCode = int64(resp.StatusCode)
		event.ResponseProto = resp.Proto
		// Save a snapshot of the response body
		var data []byte
		data, err = readSnap(&resp.Body, snapSize, t.readAll)
		if err != nil {
			t.readAllErrs.Add(1)
			resp = nil // this is how net/http likes it
		} else {
			event.ResponseBodySnap = data
		}
	}
	root.Handler.OnMeasurement(modelx.Measurement{
		HTTPRoundTripDone: event,
	})
	return resp, err
}

// CloseIdleConnections closes the idle connections.
func (t *TraceTripper) CloseIdleConnections() {
	// Adapted from net/http code
	type closeIdler interface {
		CloseIdleConnections()
	}
	if tr, ok := t.roundTripper.(closeIdler); ok {
		tr.CloseIdleConnections()
	}
}
