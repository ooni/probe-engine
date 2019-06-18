// Package netx contains network extensions.
package netx

import (
	"context"
	"net"
	"time"

	"github.com/ooni/probe-engine/httpx/retryx"
)

// RetryingDialer is a dialer where we retry dialing for
// a fixed number of attempts to increase reliability.
type RetryingDialer struct {
	// Dialer is the embedded dialer.
	net.Dialer
}

// DialContext will dial for a specific network and address
// using the specified context.
func (rd *RetryingDialer) DialContext(
	ctx context.Context, network, address string,
) (conn net.Conn, err error) {
	err = retryx.Do(ctx, func() error {
		timedctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		conn, err = rd.Dialer.DialContext(timedctx, network, address)
		return err
	})
	return
}
