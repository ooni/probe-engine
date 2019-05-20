// Package netx contains network extensions.
package netx

import (
	"context"
	"net"
)

const (
	// maxRetries is the maximum number of retries
	maxRetries = 3
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
	for i := 0; i < maxRetries; i++ {
		conn, err = rd.Dialer.DialContext(ctx, network, address)
		if err == nil {
			break
		}
	}
	return conn, err
}
