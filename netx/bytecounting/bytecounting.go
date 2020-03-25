// Package bytecounting adds bytecounting to measurable objects.
package bytecounting

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/measurable"
)

// TODO(bassosimone): we can estimate the bytes required by LookupHost
// as well using an algorithm similar to MK's one.

// Counter is the byte counting counter.
type Counter struct {
	BytesSent *atomicx.Int64
	BytesRecv *atomicx.Int64
}

// NewCounter creates a new instance of counter.
func NewCounter() *Counter {
	return &Counter{
		BytesRecv: atomicx.NewInt64(),
		BytesSent: atomicx.NewInt64(),
	}
}

// Connector is a byte counting connector.
type Connector struct {
	measurable.Connector
	*Counter
}

// DialContext implements measurable.Connector.DialContext
func (d Connector) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.Connector.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return &connWrapper{Conn: conn, Counter: d.Counter}, err
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

// WithCounter creates a copy of the provided context that is configured
// to use the specified byte counter to count bytes sent/received.
func WithCounter(ctx context.Context, counter *Counter) context.Context {
	config := measurable.ContextConfigOrDefault(ctx)
	return measurable.WithConfig(ctx, &measurable.Config{
		Connector: &Connector{
			Connector: config.Connector,
			Counter:   counter,
		},
		HTTPTransport: config.HTTPTransport,
		Resolver:      config.Resolver,
		TLSHandshaker: config.TLSHandshaker,
	})
}
