// Package httptracex contains httptrace extensions. We use these
// extensions in OONI for two purposes:
//
// 1. to emit more precise logging during normal operations, using
// the code in here combined with the httplog package;
//
// 2. to perform network measurements, as the code in here allows to
// collect more information of what happens when fetching an URL.
package httptracex

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
)

// Handler handles HTTP events.
type Handler interface {
	// RoundTripStart is when the round trip started.
	RoundTripStart(request *http.Request)

	// DNSStart is called when we start name resolution.
	DNSStart(host string)

	// DNSDone is called after name resolution.
	DNSDone(addrs []net.IPAddr, err error)

	// ConnectStart is called when we start connecting.
	ConnectStart(network, addr string)

	// ConnectDone is called after connect.
	ConnectDone(network, addr string, err error)

	// TLSHandshakeStart is called when we start the TLS handshake.
	TLSHandshakeStart()

	// TLSHandshakeDone is called after the TLS handshake.
	TLSHandshakeDone(state tls.ConnectionState, err error)

	// ConnectionReady is called when a connection is ready to be used.
	ConnectionReady(conn net.Conn, request *http.Request)

	// WroteHeaderField is called when a header field is written.
	WroteHeaderField(key string, values []string)

	// WroteHeaders is called when all headers are written.
	WroteHeaders(request *http.Request)

	// RequestBodyReadComplete is called after we've read a piece of
	// the request body from the input file.
	RequestBodyReadComplete(n int, err error)

	// RequestBodyClose is called after we've closed the body.
	RequestBodyClose(err error)

	// WroteRequest is called after the request has been written.
	WroteRequest(err error)

	// GotFirstResponseByte is called when we start reading the response.
	GotFirstResponseByte()

	// GotHeaders is called when we've got the response headers.
	GotHeaders(response *http.Response)

	// ResponseBodyReadComplete is called after we've read a piece of
	// the response body from the underlying connection.
	ResponseBodyReadComplete(n int, err error)

	// ResponseBodyClose is called after we've closed the body.
	ResponseBodyClose(err error)
}

// Measurer is an extended http.RoundTripper.
type Measurer struct {
	// RoundTripper is the child http.RoundTripper.
	http.RoundTripper

	// NewHandler creates a new handler.
	NewHandler func(ctx context.Context) Handler
}

func (m *Measurer) addTracer(
	request *http.Request, handler Handler,
) *http.Request {
	tracer := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			handler.RoundTripStart(request)
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			handler.DNSStart(info.Host)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			handler.DNSDone(info.Addrs, info.Err)
		},
		ConnectStart:      handler.ConnectStart,
		ConnectDone:       handler.ConnectDone,
		TLSHandshakeStart: handler.TLSHandshakeStart,
		TLSHandshakeDone:  handler.TLSHandshakeDone,
		GotConn: func(info httptrace.GotConnInfo) {
			handler.ConnectionReady(info.Conn, request)
		},
		WroteHeaderField: handler.WroteHeaderField,
		WroteHeaders: func() {
			handler.WroteHeaders(request)
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			handler.WroteRequest(info.Err)
		},
		GotFirstResponseByte: handler.GotFirstResponseByte,
	}
	return request.WithContext(
		httptrace.WithClientTrace(request.Context(), tracer),
	)
}

type bodyWrapper struct {
	onRead  func(n int, err error)
	onClose func(err error)
	body    io.ReadCloser
}

// Read implements the io.Reader interface for bodyWrapper
func (bw *bodyWrapper) Read(p []byte) (n int, err error) {
	n, err = bw.body.Read(p)
	bw.onRead(n, err)
	return
}

// Close implements the io.Closer interface for bodyWrapper
func (bw *bodyWrapper) Close() (err error) {
	err = bw.body.Close()
	bw.onClose(err)
	return
}

// RoundTrip performs an HTTP round trip.
func (m *Measurer) RoundTrip(request *http.Request) (*http.Response, error) {
	var handler Handler
	if m.NewHandler != nil {
		handler = m.NewHandler(request.Context())
	}
	if handler != nil {
		request = m.addTracer(request, handler)
		if request.Body != nil {
			request.Body = &bodyWrapper{
				onRead:  handler.RequestBodyReadComplete,
				onClose: handler.RequestBodyClose,
				body:    request.Body,
			}
		}
	}
	response, err := m.RoundTripper.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler.GotHeaders(response)
		// "The http Client and Transport guarantee that Body is always
		//  non-nil, even on responses without a body or responses with
		//  a zero-length body." (from the docs)
		response.Body = &bodyWrapper{
			onRead:  handler.ResponseBodyReadComplete,
			onClose: handler.ResponseBodyClose,
			body:    response.Body,
		}
	}
	return response, nil
}
