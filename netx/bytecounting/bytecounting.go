// Package bytecounting adds bytecounting to measurable objects.
package bytecounting

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/netx/measurable"
)

// TODO(bassosimone): we can estimate the bytes required by LookupHost
// as well using an algorithm similar to MK's one.

// Counter adds byte counting to measurable operations.
type Counter struct {
	measurable.Operations
	BytesSent *atomicx.Int64
	BytesRecv *atomicx.Int64
}

// NewCounter creates a new instance of counter.
func NewCounter(ops measurable.Operations) *Counter {
	return &Counter{
		Operations: ops,
		BytesRecv:  atomicx.NewInt64(),
		BytesSent:  atomicx.NewInt64(),
	}
}

// DialContext implements measurable.Connector.DialContext
func (c *Counter) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := c.Operations.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return &connWrapper{Conn: conn, Counter: c}, err
}

type connWrapper struct {
	net.Conn
	*Counter
}

func (c connWrapper) Read(p []byte) (int, error) {
	count, err := c.Conn.Read(p)
	c.Counter.BytesRecv.Add(int64(count))
	return count, err
}

func (c connWrapper) Write(p []byte) (int, error) {
	count, err := c.Conn.Write(p)
	c.Counter.BytesSent.Add(int64(count))
	return count, err
}
