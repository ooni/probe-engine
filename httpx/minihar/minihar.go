// Package minihar implements HTTP event measurements. In OONI, we use this
// functionality to store data about requests.
package minihar

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ConnectInfo contains info about connecting.
type ConnectInfo struct {
	// StartTime is when we started connecting.
	StartTime time.Time

	// EndTime is when connect returned.
	EndTime time.Time

	// Error contains the result of connect.
	Error error
}

// ReadInfo contains info about a read operation.
type ReadInfo struct {
	// Time is when the operation returned.
	Time time.Time

	// Count is the number of bytes read.
	Count int

	// Error is the result of the operation.
	Error error
}

// RoundTripSaver is a httptracex.Handler that logs events.
type RoundTripSaver struct {
	// RoundTripStartTime is when the round trip started.
	RoundTripStartTime time.Time

	// RequestMethod is the request method.
	RequestMethod string

	// RequestURL is the request full URL.
	RequestURL *url.URL

	// DNSStartTime is when we started the DNS query.
	DNSStartTime time.Time

	// DNSHostName is the hostname to resolve.
	DNSHostName string

	// DNSEndTime is when we received the DNS response.
	DNSEndTime time.Time

	// DNSAddresses contains the resolved addresses.
	DNSAddresses []net.IPAddr

	// DNSError contains the result of the DNS lookup.
	DNSError error

	// Connects contains info about the connects.
	Connects map[string]ConnectInfo

	// connectsMutex synchronizes access to Connects.
	connectsMutex sync.Mutex

	// TLSHandshakeStartTime is when we started the TLS handshake.
	TLSHandshakeStartTime time.Time

	// TLSHandshakeEndTime is when the handshake was done.
	TLSHandshakeEndTime time.Time

	// TLSConnectionState is the TLS connection state.
	TLSConnectionState *tls.ConnectionState

	// TLSHandshakeError is the result of the TLS handshake.
	TLSHandshakeError error

	// ConnectionReadyTime is when the connection was ready to use.
	ConnectionReadyTime time.Time

	// RequestURLPath is the request URL path.
	RequestURLPath string

	// RequestProto is the request protocol.
	RequestProto string

	// RequestHeaders contains the request headers.
	RequestHeaders http.Header

	// RequestHeadersWrittenTime is when we've written headers.
	RequestHeadersWrittenTime time.Time

	// RequestBodyReadInfo contains info about reading the request body
	// from a file-like when sending it to the server.
	RequestBodyReadInfo []ReadInfo

	// RequestBodyCloseTime is when we're finished reading the body.
	RequestBodyCloseTime time.Time

	// RequestBodyCloseError is the result of closing the body.
	RequestBodyCloseError error

	// RequestSentTime is when we've finished sending the request body.
	RequestSentTime time.Time

	// RequestSendError is the result of sending the request.
	RequestSendError error

	// ResponseFirstByteTime is when we receive the first response byte.
	ResponseFirstByteTime time.Time

	// ResponseHeadersTime is when we've got the response headers.
	ResponseHeadersTime time.Time

	// ResponseProto is the response protocol.
	ResponseProto string

	// ResponseStatus is the response status string (e.g. "200 Ok")
	ResponseStatus string

	// ResponseStatusCode is the response status code.
	ResponseStatusCode int

	// ResponseHeaders contains the response headers.
	ResponseHeaders http.Header

	// ResponseBodyReadInfo contains info about reading the response body
	// from the socket when receiving it from the server.
	ResponseBodyReadInfo []ReadInfo

	// ResponseBodyCloseTime is when we're finished reading the body.
	ResponseBodyCloseTime time.Time

	// ResponseBodyCloseError is the result of closing the body.
	ResponseBodyCloseError error
}

func newRoundTripSaver() *RoundTripSaver {
	return &RoundTripSaver{
		Connects:        make(map[string]ConnectInfo),
		RequestHeaders:  http.Header{},
		ResponseHeaders: http.Header{},
	}
}

// RoundTripStart is when the round trip started.
func (rts *RoundTripSaver) RoundTripStart(request *http.Request) {
	rts.RoundTripStartTime = time.Now()
	rts.RequestURL = request.URL
	rts.RequestMethod = request.Method
}

// DNSStart is called when we start name resolution.
func (rts *RoundTripSaver) DNSStart(host string) {
	rts.DNSStartTime = time.Now()
	rts.DNSHostName = host
}

// DNSDone is called after name resolution.
func (rts *RoundTripSaver) DNSDone(addrs []net.IPAddr, err error) {
	rts.DNSEndTime = time.Now()
	rts.DNSAddresses = addrs
	rts.DNSError = err
}

func (rts *RoundTripSaver) endpoint(network, addr string) string {
	return fmt.Sprintf("%s://%s", network, addr)
}

// ConnectStart is called when we start connecting.
func (rts *RoundTripSaver) ConnectStart(network, addr string) {
	epnt := rts.endpoint(network, addr)
	rts.connectsMutex.Lock()
	defer rts.connectsMutex.Unlock()
	rts.Connects[epnt] = ConnectInfo{
		StartTime: time.Now(),
	}
}

// ConnectDone is called after connect.
func (rts *RoundTripSaver) ConnectDone(network, addr string, err error) {
	epnt := rts.endpoint(network, addr)
	rts.connectsMutex.Lock()
	defer rts.connectsMutex.Unlock()
	rts.Connects[epnt] = ConnectInfo{
		EndTime: time.Now(),
		Error:   err,
	}
}

// TLSHandshakeStart is called when we start the TLS handshake.
func (rts *RoundTripSaver) TLSHandshakeStart() {
	rts.TLSHandshakeStartTime = time.Now()
}

// TLSHandshakeDone is called after the TLS handshake.
func (rts *RoundTripSaver) TLSHandshakeDone(
	state tls.ConnectionState, err error,
) {
	rts.TLSHandshakeEndTime = time.Now()
	rts.TLSConnectionState = &state
	rts.TLSHandshakeError = err
}

// ConnectionReady is called when a connection is ready to be used.
func (rts *RoundTripSaver) ConnectionReady(conn net.Conn, request *http.Request) {
	rts.ConnectionReadyTime = time.Now()
	rts.RequestMethod = request.Method
	rts.RequestURLPath = request.URL.RequestURI()
	// A connection is HTTP/2 if it's using TLS and ALPN was used. We cannot
	// rely on the Proto field because it's empty during redirects (and the
	// doc is clear that this field is not managed by clients).
	tlsconn, _ := conn.(*tls.Conn)
	if tlsconn == nil || tlsconn.ConnectionState().NegotiatedProtocol != "h2" {
		rts.RequestProto = "HTTP/1.1"
	} else {
		rts.RequestProto = "HTTP/2"
	}
}

// WroteHeaderField is called when a header field is written.
func (rts *RoundTripSaver) WroteHeaderField(key string, values []string) {
	rts.RequestHeaders[key] = values
}

// WroteHeaders is called when all headers are written.
func (rts *RoundTripSaver) WroteHeaders(request *http.Request) {
	rts.RequestHeadersWrittenTime = time.Now()
}

// RequestBodyReadComplete is called after we've read a piece of
// the request body from the input file.
func (rts *RoundTripSaver) RequestBodyReadComplete(count int, err error) {
	rts.RequestBodyReadInfo = append(rts.RequestBodyReadInfo, ReadInfo{
		Time:  time.Now(),
		Count: count,
		Error: err,
	})
}

// RequestBodyClose is called after we've closed the body.
func (rts *RoundTripSaver) RequestBodyClose(err error) {
	rts.RequestBodyCloseTime = time.Now()
	rts.RequestBodyCloseError = err
}

// WroteRequest is called after the request has been written.
func (rts *RoundTripSaver) WroteRequest(err error) {
	rts.RequestSentTime = time.Now()
	rts.RequestSendError = err
}

// GotFirstResponseByte is called when we start reading the response.
func (rts *RoundTripSaver) GotFirstResponseByte() {
	rts.ResponseFirstByteTime = time.Now()
}

// GotHeaders is called when we've got the response headers.
func (rts *RoundTripSaver) GotHeaders(response *http.Response) {
	rts.ResponseHeadersTime = time.Now()
	http2 := response.Proto == "HTTP/2" || response.Proto == "HTTP/2.0"
	if !http2 {
		rts.ResponseProto = response.Proto
	} else {
		rts.ResponseProto = "HTTP/2"
	}
	rts.ResponseStatus = response.Status
	rts.ResponseStatusCode = response.StatusCode
	rts.ResponseHeaders = response.Header
}

// ResponseBodyReadComplete is called after we've read a piece of
// the response body from the underlying connection.
func (rts *RoundTripSaver) ResponseBodyReadComplete(count int, err error) {
	rts.ResponseBodyReadInfo = append(rts.ResponseBodyReadInfo, ReadInfo{
		Time:  time.Now(),
		Count: count,
		Error: err,
	})
}

// ResponseBodyClose is called after we've closed the body.
func (rts *RoundTripSaver) ResponseBodyClose(err error) {
	rts.ResponseBodyCloseTime = time.Now()
	rts.ResponseBodyCloseError = err
}

// RequestSaver saves info about a request chain.
type RequestSaver struct {
	// CreationTime is when this saver has been created.
	CreationTime time.Time

	// RoundTrips contains the saved round trips.
	RoundTrips []*RoundTripSaver

	// mutex keeps RoundTrips access atomic.
	mutex sync.Mutex
}

// NewRoundTripSaver creates a new round trip saver and stores into
// into the chain of round-trips saved by this request saver.
func (rs *RequestSaver) NewRoundTripSaver() *RoundTripSaver {
	rts := newRoundTripSaver()
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	rs.RoundTrips = append(rs.RoundTrips, rts)
	return rts
}

type contextKey struct{}

// WithRequestSaver returns a copy of the original context where we
// have configured a request saver for all requests, as well as the
// request saver that has just been configured.
func WithRequestSaver(ctx context.Context) (context.Context, *RequestSaver) {
	rs := &RequestSaver{
		CreationTime: time.Now(),
	}
	ctx = context.WithValue(ctx, contextKey{}, rs)
	return ctx, rs
}

// ContextRequestSaver returns the request saver bound to this context or
// nil if no request saver has been bound to the saver.
func ContextRequestSaver(ctx context.Context) (rs *RequestSaver) {
	rs, _ = ctx.Value(contextKey{}).(*RequestSaver)
	return rs
}
