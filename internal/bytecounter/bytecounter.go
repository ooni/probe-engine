// Package bytecounter contains bytes counters. We want to independently
// account the bytes consumed by a session and the bytes consumed by an
// experiment. To account for bytes, you need to insert the Dialer defined
// by this package into the dialing chain. Then, you need to tell such
// dialer what is the session counter and what is the experiment counter
// using the context. Usage of the context here seems to be a necessary
// evil, because we use the same persistent connection to ps.ooni.io both
// for experiment and session traffic. We use two explicitly distinct
// counters because the alternative is that we have a single API allowing
// us to register a counter that combines new counter with previous
// counters. With such alternative design, we could mistakenly register
// the same counter more than once, possibly.
package bytecounter

import (
	"context"
	"fmt"
	"net"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/dialer"
)

// Counter is the bytes counter. You should have an instance of this
// struct per experiment and one in the session.
type Counter struct {
	Received *atomicx.Int64
	Sent     *atomicx.Int64
}

// String returns a string representation of the counter
func (c *Counter) String() string {
	return fmt.Sprintf("received: %d; sent: %d", c.Received.Load(), c.Sent.Load())
}

// New creates a new bytes counter
func New() *Counter {
	return &Counter{Received: atomicx.NewInt64(), Sent: atomicx.NewInt64()}
}

// Dialer is a byte-counting-aware dialer. To perform byte counting, you
// should make sure that you insert this dialer in the dialing chain.
type Dialer struct {
	dialer.Dialer
}

// DialContext implements dialer.Dialer.DialContext
func (d Dialer) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	exp := ContextExperimentByteCounter(ctx)
	sess := ContextSessionByteCounter(ctx)
	if exp == nil && sess == nil {
		return conn, nil // no point in wrapping
	}
	return connWrapper{Conn: conn, exp: exp, sess: sess}, nil
}

type sessionKey struct{}

// ContextSessionByteCounter retrieves the session byte counter from the context
func ContextSessionByteCounter(ctx context.Context) *Counter {
	counter, _ := ctx.Value(sessionKey{}).(*Counter)
	return counter
}

// WithSessionByteCounter assigns the session byte counter to the context
func WithSessionByteCounter(ctx context.Context, counter *Counter) context.Context {
	return context.WithValue(ctx, sessionKey{}, counter)
}

type experimentKey struct{}

// ContextExperimentByteCounter retrieves the experiment byte counter from the context
func ContextExperimentByteCounter(ctx context.Context) *Counter {
	counter, _ := ctx.Value(experimentKey{}).(*Counter)
	return counter
}

// WithExperimentByteCounter assigns the experiment byte counter to the context
func WithExperimentByteCounter(ctx context.Context, counter *Counter) context.Context {
	return context.WithValue(ctx, experimentKey{}, counter)
}

type connWrapper struct {
	net.Conn
	exp  *Counter
	sess *Counter
}

func (c connWrapper) Read(p []byte) (int, error) {
	count, err := c.Conn.Read(p)
	if c.exp != nil {
		c.exp.Received.Add(int64(count))
	}
	if c.sess != nil {
		c.sess.Received.Add(int64(count))
	}
	return count, err
}

func (c connWrapper) Write(p []byte) (int, error) {
	count, err := c.Conn.Write(p)
	if c.exp != nil {
		c.exp.Sent.Add(int64(count))
	}
	if c.sess != nil {
		c.sess.Sent.Add(int64(count))
	}
	return count, err
}
