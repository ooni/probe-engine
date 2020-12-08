// +build go1.15

package dialer

import (
	"crypto/tls"

	"github.com/lucas-clemente/quic-go"
)

// ConnectionState returns the connection state of a session which is only accessible using go 1.15
func ConnectionState(sess quic.EarlySession) tls.ConnectionState {
	return sess.ConnectionState().ConnectionState
}
