package quicdialer

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/lucas-clemente/quic-go"
)

// TLSVerifier is a Dialer used for custom verification of TLS certificates
type TLSVerifier struct {
	Dialer ContextDialer
}

// DialContext implements ContextDialer.DialContext
func (h TLSVerifier) DialContext(ctx context.Context, network string,
	host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	sess, err := h.Dialer.DialContext(ctx, network, host, tlsCfg, cfg)
	if err != nil {
		return nil, err
	}
	onlyhost, _, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}
	state := ConnectionState(sess)
	if len(state.PeerCertificates) > 0 {
		// The first element is the leaf certificate that the connection is verified against.
		err = state.PeerCertificates[0].VerifyHostname(onlyhost)
		// fmt.Println(err)
		if err == nil {
			// only succeeds if the verification was successful
			return sess, nil
		}
	}
	return nil, err

}
