// +build !go1.15

package dialer

import (
	"crypto/tls"

	"github.com/lucas-clemente/quic-go"
)

// ConnectionState returns an empty connection state
// because the quic-go interface go 1.15
func ConnectionState(sess quic.EarlySession) tls.ConnectionState {
	return tls.ConnectionState{}
}
