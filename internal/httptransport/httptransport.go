// Package httptransport contains the Transport implementation.
package httptransport

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/ooni/probe-engine/internal/errwrapper"
)

// Dialer is what an Transport expects from a dialer.
type Dialer interface {
	// DialContext is like net.Dialer.DialContext. It should split the
	// provided address using net.SplitHostPort, to get a domain name to
	// resolve. It should use some resolving functionality to map such
	// domain name to a list of IP addresses. It should then attempt to
	// dial each of them until one returns success or they all fail.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// TLSDialer is what an Transport expects from a TLS dialer.
type TLSDialer interface {
	// DialTLSContext is like net.Dialer.DialContext except that it also
	// establishes a TLS connection. By default the SNI is extracted from
	// the provided address by using net.SplitHostPort.
	DialTLSContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Transport is the common interface of all transports.
type Transport interface {
	// RoundTrip performs the specified HTTP request and returns either
	// a response or an error. See net/http.RoundTripper.RoundTrip.
	RoundTrip(req *http.Request) (*http.Response, error)

	// CloseIdleConnections closes the idle connections, if any.
	CloseIdleConnections()
}

// NewBase creates a new instance of the base HTTP transport. The base transport
// is a clone of the default http.Transport where we use the provide dialers to
// establish new connections, and where we configure settings that help us to run
// measurements and observe less noisy output.
func NewBase(dialer Dialer, tlsDialer TLSDialer) *http.Transport {
	txp := http.DefaultTransport.(*http.Transport).Clone()
	txp.DialContext = dialer.DialContext
	txp.DialTLSContext = tlsDialer.DialTLSContext
	txp.DisableCompression = true // we want to see all headers
	txp.MaxConnsPerHost = 1       // make events less noisy
	return txp
}

// Logger is the logger interface assumed by this package
type Logger interface {
	Debugf(format string, v ...interface{})
}

// The Logging transport is a transport that implements logging. It will
// specifically log the beginning and end of the round trip.
type Logging struct {
	Transport
	Logger Logger
}

// RoundTrip implements Transport.RoundTrip.
func (txp Logging) RoundTrip(req *http.Request) (*http.Response, error) {
	txp.Logger.Debugf("> %s %s", req.Method, req.URL)
	resp, err := txp.Transport.RoundTrip(req)
	if err != nil {
		txp.Logger.Debugf("< %s", err.Error())
		return nil, err
	}
	txp.Logger.Debugf("< %d", resp.StatusCode)
	return resp, nil
}

// ErrWrapper is a transport that wraps errors as OONI errors.
type ErrWrapper struct {
	Transport
}

// RoundTrip implements Transport.RoundTrip.
func (txp ErrWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := txp.Transport.RoundTrip(req)
	err = errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: "http_round_trip",
	}.MaybeBuild()
	return resp, err
}

// HeaderAdder is a transport that adds some headers that we
// always want to set explicitly in the request.
type HeaderAdder struct {
	Transport
	UserAgent string
}

// RoundTrip implements Transport.RoundTrip.
func (txp HeaderAdder) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		if txp.UserAgent != "" {
			req.Header.Set("User-Agent", txp.UserAgent)
		} else {
			req.Header["User-Agent"] = nil // disable sending user agent
		}
	}
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	req.Header.Set("Host", host) // have it set explicitly
	return txp.Transport.RoundTrip(req)
}

// SnapshotSaver saves a snapshot of the response body
type SnapshotSaver struct {
	Transport
	SnapshotSize int64
	snapshots    []BodySnapshot
	mu           sync.Mutex
}

// Snapshots returns the saved body snapshots
func (txp *SnapshotSaver) Snapshots() []BodySnapshot {
	txp.mu.Lock()
	snapshots := txp.snapshots
	txp.snapshots = nil
	txp.mu.Unlock()
	return snapshots
}

// RoundTrip implements Transport.RoundTrip.
func (txp *SnapshotSaver) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := txp.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	size := txp.SnapshotSize
	if size <= 0 {
		size = 1 << 18
	}
	reader := io.LimitReader(resp.Body, size)
	data, err := ioutil.ReadAll(reader)
	snapshot := BodySnapshot{Data: string(data), Err: err, ReadCloser: resp.Body,
		Time: time.Now(), URL: req.URL.String()}
	txp.mu.Lock()
	txp.snapshots = append(txp.snapshots, snapshot)
	txp.mu.Unlock()
	resp.Body = snapshot
	return resp, nil
}

// BodySnapshot is a snapshot of the response body
type BodySnapshot struct {
	io.ReadCloser
	Data string
	Err  error
	Time time.Time
	URL  string
}

// Read implements ReadCloser.Reader.Read
func (body BodySnapshot) Read(p []byte) (int, error) {
	if body.Err != nil {
		return 0, body.Err
	}
	return body.ReadCloser.Read(p)
}

// EventsSaver saves events occurring during the round trip
type EventsSaver struct {
	Transport
	events []Events
	mu     sync.Mutex
}

// RoundTrip implements Transport.RoundTrip.
func (txp *EventsSaver) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		rte Events
		mu  sync.Mutex
	)
	tracer := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			mu.Lock()
			rte.GetConnTime = time.Now()
			rte.GetConnAddress = hostPort
			mu.Unlock()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			mu.Lock()
			rte.GotConnTime = time.Now()
			rte.GotConnAddress = info.Conn.LocalAddr().String()
			mu.Unlock()
		},
		GotFirstResponseByte: func() {
			mu.Lock()
			rte.GotFirstResponseByteTime = time.Now()
			mu.Unlock()
		},
		WroteHeaders: func() {
			mu.Lock()
			rte.WroteHeadersTime = time.Now()
			mu.Unlock()
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			mu.Lock()
			rte.WroteRequestTime = time.Now()
			mu.Unlock()
		},
	}
	rte.Method = req.Method
	rte.URL = req.URL.String()
	rte.RequestHeaders = req.Header
	rte.RoundTripStartTime = time.Now()
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), tracer))
	resp, err := txp.Transport.RoundTrip(req)
	rte.RoundTripEndTime = time.Now()
	if resp != nil {
		rte.ResponseHeaders = resp.Header
		rte.StatusCode = resp.StatusCode
	}
	txp.mu.Lock()
	txp.events = append(txp.events, rte)
	txp.mu.Unlock()
	return resp, err
}

// ReadEvents reads the saved events and returns them
func (txp *EventsSaver) ReadEvents() []Events {
	txp.mu.Lock()
	events := txp.events
	txp.events = nil
	txp.mu.Unlock()
	return events
}

// Events describes round trip events
type Events struct {
	Method                   string
	RequestHeaders           http.Header
	URL                      string
	RoundTripStartTime       time.Time
	GetConnTime              time.Time
	GetConnAddress           string
	GotConnAddress           string
	GotConnTime              time.Time
	GotFirstResponseByteTime time.Time
	WroteHeadersTime         time.Time
	WroteRequestTime         time.Time
	RoundTripEndTime         time.Time
	StatusCode               int
	ResponseHeaders          http.Header
}
