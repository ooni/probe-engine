package httptransport

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/ooni/probe-engine/netx/dialer"
)

// HTTP3Dialer is the definition of dialer for HTTP3 transport assumed by this package.
type HTTP3Dialer interface {
	Dial(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

// HTTP3ContextDialer is a dialer for HTTP3 transport using Context.
type HTTP3ContextDialer interface {
	DialContext(ctx context.Context, network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

// HTTP3WrapperDialer is a HTTP3Dialer that wraps a ContextDialer
// This is necessary because the http3 RoundTripper does not support a DialContext method.
type HTTP3WrapperDialer struct {
	Dialer HTTP3ContextDialer
}

// Dial implements HTTP3Dialer.Dial
func (d HTTP3WrapperDialer) Dial(network, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	return d.Dialer.DialContext(context.Background(), network, host, tlsCfg, cfg)
}

// HTTP3Transport is a httptransport.RoundTripper using the http3 protocol.
type HTTP3Transport struct {
	http3.RoundTripper
}

// CloseIdleConnections closes all the connections opened by this transport.
func (t *HTTP3Transport) CloseIdleConnections() {
	// TODO(kelmenhorst): implement
}

// NewHTTP3Transport creates a new HTTP3Transport instance.
func NewHTTP3Transport(config Config) RoundTripper {
	txp := &HTTP3Transport{}
	txp.QuicConfig = &quic.Config{}
	if tlsdialer, ok := config.TLSDialer.(dialer.TLSDialer); ok {
		txp.TLSClientConfig = tlsdialer.Config
	}
	txp.Dial = config.HTTP3Dialer.Dial
	return txp
}

var _ RoundTripper = &http.Transport{}
