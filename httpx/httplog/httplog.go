// Package httplog implements HTTP event logging. In OONI, we use this
// functionality to emit pleasant logging during normal operations.
package httplog

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"github.com/ooni/probe-engine/log"
)

var (
	tlsVersion = map[uint16]string{
		tls.VersionSSL30: "SSLv3",
		tls.VersionTLS10: "TLSv1",
		tls.VersionTLS11: "TLSv1.1",
		tls.VersionTLS12: "TLSv1.2",
		tls.VersionTLS13: "TLSv1.3",
	}
)

// RoundTripLogger is a httptracex.Handler that logs events.
type RoundTripLogger struct {
	// Logger is the logs emitter.
	Logger log.Logger

	// header contains the emitted headers.
	headers http.Header
}

// DNSStart is called when we start name resolution.
func (rtl *RoundTripLogger) DNSStart(host string) {
	rtl.Logger.Debugf("dns: resolving %s", host)
}

// DNSDone is called after name resolution.
func (rtl *RoundTripLogger) DNSDone(addrs []net.IPAddr, err error) {
	if err != nil {
		rtl.Logger.Debugf("dns: error: %s", err.Error())
		return
	}
	rtl.Logger.Debugf("dns: got %d entries", len(addrs))
	for _, addr := range addrs {
		rtl.Logger.Debugf("- %s", addr.String())
	}
}

// ConnectStart is called when we start connecting.
func (rtl *RoundTripLogger) ConnectStart(network, addr string) {
	rtl.Logger.Debugf("connect: using %s, %s", network, addr)
}

// ConnectDone is called after connect.
func (rtl *RoundTripLogger) ConnectDone(network, addr string, err error) {
	if err != nil {
		rtl.Logger.Debugf("connect: error: %s", err.Error())
		return
	}
	rtl.Logger.Debugf("connect: connected to %s, %s", network, addr)
}

// TLSHandshakeStart is called when we start the TLS handshake.
func (rtl *RoundTripLogger) TLSHandshakeStart() {
	rtl.Logger.Debug("tls: starting handshake")
}

// TLSHandshakeDone is called after the TLS handshake.
func (rtl *RoundTripLogger) TLSHandshakeDone(
	state tls.ConnectionState, err error,
) {
	if err != nil {
		rtl.Logger.Debugf("tls: handshake error: %s", err.Error())
		return
	}
	rtl.Logger.Debug("tls: handshake OK")
	rtl.Logger.Debugf("- negotiated protocol: %s", state.NegotiatedProtocol)
	rtl.Logger.Debugf("- version: %s", tlsVersion[state.Version])
}

// ConnectionReady is called when a connection is ready to be used.
func (rtl *RoundTripLogger) ConnectionReady(conn net.Conn) {
	rtl.Logger.Debugf(
		"http: connection to %s ready; sending request", conn.RemoteAddr(),
	)
	rtl.headers = make(http.Header, 16) // reset
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
	for _, value := range values {
		rtl.headers.Add(key, value)
	}
}

// WroteHeaders is called when all headers are written.
func (rtl *RoundTripLogger) WroteHeaders(request *http.Request) {
	http2 := rtl.headers.Get(":method") != ""
	if !http2 {
		rtl.Logger.Debugf(
			"> %s %s HTTP/1.1", request.Method, request.URL.RequestURI(),
		)
	} else {
		for _, s := range []string{":method", ":scheme", ":authority", ":path"} {
			rtl.logSingleHeader(http2, ">", s, rtl.headers.Get(s))
		}
	}
	for key, values := range rtl.headers {
		if strings.HasPrefix(key, ":") {
			continue
		}
		rtl.logHeaderVector(http2, ">", key, values)
	}
	rtl.Logger.Debug(">")
}

func (rtl *RoundTripLogger) onReadComplete(
	n int, err error, dir, what string,
) {
	if n > 0 {
		rtl.Logger.Debugf("%s [%d bytes data]", dir, n)
	}
	if err != nil {
		rtl.Logger.Debugf("http: reading %s body: %s", what, err.Error())
	}
}

// RequestBodyReadComplete is called after we've read a piece of
// the request body from the underlying connection.
func (rtl *RoundTripLogger) RequestBodyReadComplete(n int, err error) {
	rtl.onReadComplete(n, err, "}", "request")
}

func (rtl *RoundTripLogger) onClose(err error, what string) {
	if err != nil {
		rtl.Logger.Debugf("http: closing %s body failed: %s", what, err.Error())
		return
	}
	rtl.Logger.Debugf("http: closed %s body", what)
}

// RequestBodyClose is called after we've closed the body.
func (rtl *RoundTripLogger) RequestBodyClose(err error) {
	rtl.onClose(err, "request")
}

// WroteRequest is called after the request has been written.
func (rtl *RoundTripLogger) WroteRequest(err error) {
	if err != nil {
		rtl.Logger.Debugf("http: sending request failed: %s", err.Error())
		return
	}
	rtl.Logger.Debugf("http: request sent; waiting for response")
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
