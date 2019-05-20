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
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptrace"
)

// Handler handles HTTP events.
type Handler interface {
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
	ConnectionReady(conn net.Conn)

	// WroteHeaderField is called when a header field is written.
	WroteHeaderField(key string, values []string)

	// WroteHeaders is called when all headers are written.
	WroteHeaders(request *http.Request)

	// WroteRequest is called after the request has been written.
	WroteRequest(err error)

	// GotFirstResponseByte is called when we start reading the response.
	GotFirstResponseByte()

	// GotHeaders is called when we've got the response headers.
	GotHeaders(response *http.Response)
}

// Measurer is an extended http.RoundTripper.
type Measurer struct {
	// RoundTripper is the child http.RoundTripper.
	http.RoundTripper

	// Handler is the event handler
	Handler Handler
}

func (m *Measurer) addTracer(request *http.Request) *http.Request {
	tracer := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			m.Handler.DNSStart(info.Host)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			m.Handler.DNSDone(info.Addrs, info.Err)
		},
		ConnectStart:      m.Handler.ConnectStart,
		ConnectDone:       m.Handler.ConnectDone,
		TLSHandshakeStart: m.Handler.TLSHandshakeStart,
		TLSHandshakeDone:  m.Handler.TLSHandshakeDone,
		GotConn: func(info httptrace.GotConnInfo) {
			m.Handler.ConnectionReady(info.Conn)
		},
		WroteHeaderField: m.Handler.WroteHeaderField,
		WroteHeaders: func() {
			m.Handler.WroteHeaders(request)
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			m.Handler.WroteRequest(info.Err)
		},
		GotFirstResponseByte: m.Handler.GotFirstResponseByte,
	}
	return request.WithContext(
		httptrace.WithClientTrace(request.Context(), tracer),
	)
}

// RoundTrip performs an HTTP round trip.
func (m *Measurer) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := m.RoundTripper.RoundTrip(m.addTracer(request))
	if err != nil {
		return nil, err
	}
	m.Handler.GotHeaders(response)
	return response, nil
}
