// +build !go1.15

package dialer

import (
	"crypto/tls"

	"github.com/lucas-clemente/quic-go"
)

// ConnectionState returns an empty ConnectionState because a QUIC Session's ConnectionState is only exposed using go1.15
func ConnectionState(sess quic.EarlySession) tls.ConnectionState {
	return tls.ConnectionState{}
}
