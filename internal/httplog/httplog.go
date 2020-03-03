// Package httplog implements HTTP event logging. In OONI, we use this
// functionality to emit pleasant logging during normal operations.
package httplog

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"github.com/ooni/probe-engine/internal/tlsx"
	"github.com/ooni/probe-engine/log"
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
	rtl.Logger.Debugf("- version: %s", tlsx.VersionString(state.Version))
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
