package dialer

import (
	"context"
	"net"

	"golang.org/x/net/proxy"
)

// SOCKS5Dialer is a dialer that uses SOCKS5
type SOCKS5Dialer struct {
	Address string
	Dialer
}

// socks5DialerAdapter is necessary because Go code uses duck typing
// to check whether a proxy.Dialer is a proxy.ContextDialer and in such
// case it calls dial context, so we need to wrap our child Dialer.
//
// See https://git.io/JfJ4g
type socks5DialerAdapter struct {
	Dialer
}

func (d socks5DialerAdapter) Dial(network, address string) (net.Conn, error) {
	// Implementation note: this function should never be called because
	// the code should prefer DialContext when available.
	//
	// See https://git.io/JfJ4g
	return d.Dialer.DialContext(context.Background(), network, address)
}

// DialContext implements Dialer.DialContext.
func (d SOCKS5Dialer) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	child, err := proxy.SOCKS5(
		"tcp", d.Address, nil, socks5DialerAdapter{Dialer: d.Dialer})
	if err != nil {
		return nil, err
	}
	connch := make(chan net.Conn)
	errch := make(chan error, 1)
	go func() {
		conn, err := child.Dial(network, address)
		if err != nil {
			errch <- err
			return
		}
		select {
		case connch <- conn:
		default:
			conn.Close()
		}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errch:
		return nil, err
	case conn := <-connch:
		return conn, nil
	}
}
