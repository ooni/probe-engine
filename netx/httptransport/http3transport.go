package httptransport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
)

// HTTP3Transport consists of a http3. RoundTripper
// and possibly fields which are not implemented by http3.Roundtripper, to mimic http.Transport
type HTTP3Transport struct {
	http3.RoundTripper
	DialContext        func(ctx context.Context, network, addr string) (net.Conn, error)
	DialTLSContext     func(ctx context.Context, network, addr string) (net.Conn, error)
	DisableCompression bool
	MaxConnsPerHost    int
}

// CloseIdleConnections TODO (necessary for interface compliance)
func (t *HTTP3Transport) CloseIdleConnections() {
}

// NewHTTP3Transport creates a new http3 transport.
// That is a transport using the quic-go library.
func NewHTTP3Transport(dialer Dialer, tlsDialer TLSDialer) *HTTP3Transport {
	txp := &HTTP3Transport{}
	txp.QuicConfig = new(quic.Config)
	// this is how a basic custom dialer could look like
	txp.Dial = func(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
		// fmt.Println("QUIC Dialer")
		return quic.DialAddrEarly(addr, tlsCfg, cfg)
	}
	return txp
}

var _ RoundTripper = &http.Transport{}
