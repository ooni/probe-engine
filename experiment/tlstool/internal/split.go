package internal

import (
	"bytes"
	"context"
	"net"
	"time"

	"github.com/ooni/probe-engine/netx"
)

// SplitDialer is a dialer that splits writes according to a
// pattern and may delay the second write of delay milliseconds.
type SplitDialer struct {
	netx.Dialer
	Delay   int64
	Pattern string
}

// DialContext implements netx.Dialer.DialContext.
func (d SplitDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return SplitConn{Conn: conn, Delay: d.Delay, Pattern: []byte(d.Pattern)}, nil
}

// SplitConn is a conn that splits writes.
type SplitConn struct {
	net.Conn
	BeforeSecondWrite func() // for testing
	Delay             int64
	Pattern           []byte
}

// Write implements net.Conn.Write.
func (c SplitConn) Write(b []byte) (int, error) {
	idx := bytes.Index(b, c.Pattern)
	if idx > -1 {
		idx += len(c.Pattern) / 2
		if _, err := c.Conn.Write(b[:idx]); err != nil {
			return 0, err
		}
		<-time.After(time.Duration(c.Delay) * time.Millisecond)
		if c.BeforeSecondWrite != nil {
			c.BeforeSecondWrite()
		}
		if _, err := c.Conn.Write(b[idx:]); err != nil {
			return 0, err
		}
		return len(b), nil
	}
	return c.Conn.Write(b)
}
