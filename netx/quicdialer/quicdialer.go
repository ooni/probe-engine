package quicdialer

import (
	"context"
	"crypto/tls"

	"github.com/lucas-clemente/quic-go"
)

// QUICContextDialer is a dialer for QUIC using Context.
type QUICContextDialer interface {
	DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}

// QUICDialer is the definition of a dialer that can be used for Dialing QUIC connections
type QUICDialer interface {
	Dial(network, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error)
}
