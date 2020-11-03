//+build !go1.15,!PSIPHON_DISABLE_QUIC

package httptransport

import (
	"net/http"
	"os"

	"github.com/apex/log"
)

// HTTP3Transport is a httptransport.RoundTripper using the http3 protocol.
type NilRoundTripper struct {
	RoundTripper
}

// CloseIdleConnections closes all the connections opened by this transport.
func (t *NilRoundTripper) CloseIdleConnections() {
	// do nothing, for interface compliance
}

// NewHTTP3Transport creates a new HTTP3Transport instance.
func NewHTTP3Transport(dialer Dialer, tlsDialer TLSDialer) NilRoundTripper {
	log.Errorf("%s", "HTTP3 not available. Please use Go 1.15 and the tag PSIPHON_DISABLE_QUIC")
	os.Exit(1)
	return NilRoundTripper{}
}

var _ RoundTripper = &http.Transport{}
