package dialer

import (
	"context"
	"fmt"
	"net"

	"github.com/ooni/probe-engine/atomicx"
)

// ByteCounter is the bytes counter. You should have an instance of this
// struct per experiment and another instance in the session.
type ByteCounter struct {
	Received *atomicx.Float64
	Sent     *atomicx.Float64
}

// String returns a string representation of the counter
func (c *ByteCounter) String() string {
	return fmt.Sprintf("received: %f; sent: %f", c.Received.Load(), c.Sent.Load())
}

// NewByteCounter creates a new bytes counter
func NewByteCounter() *ByteCounter {
	return &ByteCounter{Received: atomicx.NewFloat64(), Sent: atomicx.NewFloat64()}
}

// ByteCounterDialer is a byte-counting-aware dialer. To perform byte counting, you
// should make sure that you insert this dialer in the dialing chain.
//
// Bug
//
// This implementation cannot properly account for the bytes that are sent by
// persistent connections, because they strick to the counters set when the
// connection was established. This typically means we miss the bytes sent and
// received when submitting a measurement. Such bytes are specifically not
// see by the experiment specific byte counter.
//
// For this reason, this implementation may be heavily changed/removed.
type ByteCounterDialer struct {
	Dialer
}

// DialContext implements Dialer.DialContext
func (d ByteCounterDialer) DialContext(
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
	return byteCounterConnWrapper{Conn: conn, exp: exp, sess: sess}, nil
}

type byteCounterSessionKey struct{}

// ContextSessionByteCounter retrieves the session byte counter from the context
func ContextSessionByteCounter(ctx context.Context) *ByteCounter {
	counter, _ := ctx.Value(byteCounterSessionKey{}).(*ByteCounter)
	return counter
}

// WithSessionByteCounter assigns the session byte counter to the context
func WithSessionByteCounter(ctx context.Context, counter *ByteCounter) context.Context {
	return context.WithValue(ctx, byteCounterSessionKey{}, counter)
}

type byteCounterExperimentKey struct{}

// ContextExperimentByteCounter retrieves the experiment byte counter from the context
func ContextExperimentByteCounter(ctx context.Context) *ByteCounter {
	counter, _ := ctx.Value(byteCounterExperimentKey{}).(*ByteCounter)
	return counter
}

// WithExperimentByteCounter assigns the experiment byte counter to the context
func WithExperimentByteCounter(ctx context.Context, counter *ByteCounter) context.Context {
	return context.WithValue(ctx, byteCounterExperimentKey{}, counter)
}

type byteCounterConnWrapper struct {
	net.Conn
	exp  *ByteCounter
	sess *ByteCounter
}

func (c byteCounterConnWrapper) Read(p []byte) (int, error) {
	count, err := c.Conn.Read(p)
	if c.exp != nil {
		c.exp.Received.Add(float64(count) / 1024)
	}
	if c.sess != nil {
		c.sess.Received.Add(float64(count) / 1024)
	}
	return count, err
}

func (c byteCounterConnWrapper) Write(p []byte) (int, error) {
	count, err := c.Conn.Write(p)
	if c.exp != nil {
		c.exp.Sent.Add(float64(count) / 1024)
	}
	if c.sess != nil {
		c.sess.Sent.Add(float64(count) / 1024)
	}
	return count, err
}
