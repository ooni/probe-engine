package httptransport

import (
	"net/http"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
)

// HTTP3Transport is a httptransport.RoundTripper using the http3 protocol.
type HTTP3Transport struct {
	http3.RoundTripper
}

// CloseIdleConnections closes all the connections opened by this transport.
func (t *HTTP3Transport) CloseIdleConnections() {
	// TODO(kelmenhorst): implement
}

// NewHTTP3Transport creates a new HTTP3Transport instance.
func NewHTTP3Transport(dialer Dialer, tlsDialer TLSDialer) RoundTripper {
	txp := &HTTP3Transport{}
	txp.QuicConfig = &quic.Config{}
	return txp
}

var _ RoundTripper = &http.Transport{}
