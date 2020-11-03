//+build go1.15 DISABLE_QUIC

package httptransport

import (
	"crypto/tls"
	"net/http"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
)

// HTTP3Transport consists of a http3. RoundTripper
// and possibly fields which are not implemented by http3.Roundtripper, to mimic http.Transport
type HTTP3Transport struct {
	http3.RoundTripper
}

// CloseIdleConnections closes all the connections opened by this transport.
func (t *HTTP3Transport) CloseIdleConnections() {
	// TODO(kelmenhorst): implement
}

// NewHTTP3Transport creates a new http3 transport.
// That is a transport using the quic-go library.
func NewHTTP3Transport(dialer Dialer, tlsDialer TLSDialer) RoundTripper {
	txp := &HTTP3Transport{}
	txp.QuicConfig = &quic.Config{}
	txp.Dial = func(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
		// quic.DialAddrEarlyContext is not in the official release of quic-go yet
		return quic.DialAddrEarly(addr, tlsCfg, cfg)
	}
	return txp
}

var _ RoundTripper = &http.Transport{}
