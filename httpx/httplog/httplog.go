// Package httplog implements HTTP event logging. In OONI, we use this
// functionality to emit pleasant logging during normal operations.
package httplog

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ooni/probe-engine/httpx/tlsx"
	"github.com/ooni/probe-engine/log"
)

// RoundTripLogger is a httptracex.Handler that logs events.
type RoundTripLogger struct {
	// Logger is the logs emitter.
	Logger log.Logger
}

// DNSStart is called when we start name resolution.
func (rtl *RoundTripLogger) DNSStart(host string) {
	rtl.Logger.Debugf("dns: resolving %s", host)
}

func (rtl *RoundTripLogger) formatError(err error) string {
	if err != nil {
		return err.Error()
	}
	return "no error"
}

// DNSDone is called after name resolution.
func (rtl *RoundTripLogger) DNSDone(addrs []net.IPAddr, err error) {
	rtl.Logger.Debugf("dns: %s", rtl.formatError(err))
	for _, addr := range addrs {
		rtl.Logger.Debugf("- %s", addr.String())
	}
}

// ConnectStart is called when we start connecting.
func (rtl *RoundTripLogger) ConnectStart(network, addr string) {
	rtl.Logger.Debugf("connect: to %s://%s...", network, addr)
}

// ConnectDone is called after connect.
func (rtl *RoundTripLogger) ConnectDone(network, addr string, err error) {
	rtl.Logger.Debugf("connect: to %s://%s: %s", network, addr, rtl.formatError(err))
}

// TLSHandshakeStart is called when we start the TLS handshake.
func (rtl *RoundTripLogger) TLSHandshakeStart() {
	rtl.Logger.Debug("tls: starting handshake")
}

// TLSHandshakeDone is called after the TLS handshake.
func (rtl *RoundTripLogger) TLSHandshakeDone(
	state tls.ConnectionState, err error,
) {
	rtl.Logger.Debugf("tls: handshake: %s", rtl.formatError(err))
	rtl.Logger.Debugf("- negotiated protocol: %s", state.NegotiatedProtocol)
	rtl.Logger.Debugf("- version: %s", tlsx.TLSVersionString[state.Version])
	rtl.Logger.Debugf("- cipher suite: %s", tlsx.TLSCipherSuiteString[state.CipherSuite])
	for _, cert := range state.PeerCertificates {
		rtl.Logger.Debug(tlsx.TLSCertToPEM(cert))
	}
}

// ConnectionReady is called when a connection is ready to be used.
func (rtl *RoundTripLogger) ConnectionReady(conn net.Conn, request *http.Request) {
	rtl.Logger.Debugf(
		"http: connection to %s ready; sending request", conn.RemoteAddr(),
	)
	// A connection is HTTP/2 if it's using TLS and ALPN was used. We cannot
	// rely on the Proto field because it's empty during redirects (and the
	// doc is clear that this field is not managed by clients).
	tlsconn, _ := conn.(*tls.Conn)
	if tlsconn == nil || tlsconn.ConnectionState().NegotiatedProtocol != "h2" {
		rtl.Logger.Debugf(
			"> %s %s %s", request.Method, request.URL.RequestURI(), request.Proto,
		)
	}
}

func (rtl *RoundTripLogger) logSingleHeader(
	http2 bool, prefix, key, value string,
) {
	if http2 {
		key = strings.ToLower(key)
	}
	rtl.Logger.Debugf("%s %s: %s", prefix, key, value)
}

func (rtl *RoundTripLogger) logHeaderVector(
	http2 bool, prefix, key string, values []string,
) {
	for _, value := range values {
		rtl.logSingleHeader(http2, prefix, key, value)
	}
}

// WroteHeaderField is called when a header field is written.
func (rtl *RoundTripLogger) WroteHeaderField(key string, values []string) {
	const whatever = false // headers are already okay case-wise here
	rtl.logHeaderVector(whatever, ">", key, values)
}

// WroteHeaders is called when all headers are written.
func (rtl *RoundTripLogger) WroteHeaders(request *http.Request) {
	rtl.Logger.Debug(">")
}

func (rtl *RoundTripLogger) onReadComplete(
	n int, err error, dir, what string,
) {
	if n > 0 {
		rtl.Logger.Debugf("%s [%d bytes data]", dir, n)
	}
	if err != nil && err != io.EOF {
		rtl.Logger.Debugf("http: reading %s body: %s", what, err.Error())
	}
}

// RequestBodyReadComplete is called after we've read a piece of
// the request body from the input file.
func (rtl *RoundTripLogger) RequestBodyReadComplete(n int, err error) {
	rtl.onReadComplete(n, err, "}", "request")
}

func (rtl *RoundTripLogger) onClose(err error, what string) {
	rtl.Logger.Debugf("http: closing %s body: %s", what, rtl.formatError(err))
}

// RequestBodyClose is called after we've closed the body.
func (rtl *RoundTripLogger) RequestBodyClose(err error) {
	rtl.onClose(err, "request")
}

// WroteRequest is called after the request has been written.
func (rtl *RoundTripLogger) WroteRequest(err error) {
	rtl.Logger.Debugf("http: sending request: %s", rtl.formatError(err))
}

// GotFirstResponseByte is called when we start reading the response.
func (rtl *RoundTripLogger) GotFirstResponseByte() {
	rtl.Logger.Debugf("http: start receiving response")
}

// GotHeaders is called when we've got the response headers.
func (rtl *RoundTripLogger) GotHeaders(response *http.Response) {
	http2 := response.Proto == "HTTP/2" || response.Proto == "HTTP/2.0"
	if !http2 {
		rtl.Logger.Debugf("< %s %s", response.Proto, response.Status)
	} else {
		rtl.Logger.Debugf("< :status: %d", response.StatusCode)
	}
	for key, values := range response.Header {
		rtl.logHeaderVector(http2, "<", key, values)
	}
	rtl.Logger.Debug("<")
}

// ResponseBodyReadComplete is called after we've read a piece of
// the response body from the underlying connection.
func (rtl *RoundTripLogger) ResponseBodyReadComplete(n int, err error) {
	rtl.onReadComplete(n, err, "{", "response")
}

// ResponseBodyClose is called after we've closed the body.
func (rtl *RoundTripLogger) ResponseBodyClose(err error) {
	rtl.onClose(err, "response")
}
