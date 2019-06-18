package httptracex

import (
	"crypto/tls"
	"net"
	"net/http"
)

// MultiHandler is a handler that fans out to multiple childs.
type MultiHandler struct {
	// All contains all the handlers.
	All []Handler
}

// RoundTripStart is when the round trip started.
func (mh *MultiHandler) RoundTripStart(request *http.Request) {
	for _, handler := range mh.All {
		handler.RoundTripStart(request)
	}
}

// DNSStart is called when we start name resolution.
func (mh *MultiHandler) DNSStart(host string) {
	for _, handler := range mh.All {
		handler.DNSStart(host)
	}
}

// DNSDone is called after name resolution.
func (mh *MultiHandler) DNSDone(addrs []net.IPAddr, err error) {
	for _, handler := range mh.All {
		handler.DNSDone(addrs, err)
	}
}

// ConnectStart is called when we start connecting.
func (mh *MultiHandler) ConnectStart(network, addr string) {
	for _, handler := range mh.All {
		handler.ConnectStart(network, addr)
	}
}

// ConnectDone is called after connect.
func (mh *MultiHandler) ConnectDone(network, addr string, err error) {
	for _, handler := range mh.All {
		handler.ConnectDone(network, addr, err)
	}
}

// TLSHandshakeStart is called when we start the TLS handshake.
func (mh *MultiHandler) TLSHandshakeStart() {
	for _, handler := range mh.All {
		handler.TLSHandshakeStart()
	}
}

// TLSHandshakeDone is called after the TLS handshake.
func (mh *MultiHandler) TLSHandshakeDone(
	state tls.ConnectionState, err error,
) {
	for _, handler := range mh.All {
		handler.TLSHandshakeDone(state, err)
	}
}

// ConnectionReady is called when a connection is ready to be used.
func (mh *MultiHandler) ConnectionReady(conn net.Conn, request *http.Request) {
	for _, handler := range mh.All {
		handler.ConnectionReady(conn, request)
	}
}

// WroteHeaderField is called when a header field is written.
func (mh *MultiHandler) WroteHeaderField(key string, values []string) {
	for _, handler := range mh.All {
		handler.WroteHeaderField(key, values)
	}
}

// WroteHeaders is called when all headers are written.
func (mh *MultiHandler) WroteHeaders(request *http.Request) {
	for _, handler := range mh.All {
		handler.WroteHeaders(request)
	}
}

// RequestBodyReadComplete is called after we've read a piece of
// the request body from the input file.
func (mh *MultiHandler) RequestBodyReadComplete(n int, err error) {
	for _, handler := range mh.All {
		handler.RequestBodyReadComplete(n, err)
	}
}

// RequestBodyClose is called after we've closed the body.
func (mh *MultiHandler) RequestBodyClose(err error) {
	for _, handler := range mh.All {
		handler.RequestBodyClose(err)
	}
}

// WroteRequest is called after the request has been written.
func (mh *MultiHandler) WroteRequest(err error) {
	for _, handler := range mh.All {
		handler.WroteRequest(err)
	}
}

// GotFirstResponseByte is called when we start reading the response.
func (mh *MultiHandler) GotFirstResponseByte() {
	for _, handler := range mh.All {
		handler.GotFirstResponseByte()
	}
}

// GotHeaders is called when we've got the response headers.
func (mh *MultiHandler) GotHeaders(response *http.Response) {
	for _, handler := range mh.All {
		handler.GotHeaders(response)
	}
}

// ResponseBodyReadComplete is called after we've read a piece of
// the response body from the underlying connection.
func (mh *MultiHandler) ResponseBodyReadComplete(n int, err error) {
	for _, handler := range mh.All {
		handler.ResponseBodyReadComplete(n, err)
	}
}

// ResponseBodyClose is called after we've closed the body.
func (mh *MultiHandler) ResponseBodyClose(err error) {
	for _, handler := range mh.All {
		handler.ResponseBodyClose(err)
	}
}
