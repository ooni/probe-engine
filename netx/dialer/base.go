package dialer

import (
	"context"
	"net"
	"time"

	"github.com/ooni/probe-engine/netx/internal/connid"
	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// TimeoutDialer is a wrapper for the system dialer
type TimeoutDialer struct {
	modelx.Dialer
	ConnectTimeout time.Duration // default: 30 seconds
}

// DialContext implements Dialer.DialContext
func (d TimeoutDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	timeout := d.ConnectTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return d.Dialer.DialContext(ctx, network, address)
}

// ErrWrapperDialer is a dialer that performs err wrapping
type ErrWrapperDialer struct {
	modelx.Dialer
}

// DialContext implements Dialer.DialContext
func (d ErrWrapperDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialID := dialid.ContextDialID(ctx)
	conn, err := d.Dialer.DialContext(ctx, network, address)
	err = errwrapper.SafeErrWrapperBuilder{
		// ConnID does not make any sense if we've failed and the error
		// does not make any sense (and is nil) if we succeded.
		DialID:    dialID,
		Error:     err,
		Operation: "connect",
	}.MaybeBuild()
	return conn, err
}

// EmitterDialer is a dialer that emits events
type EmitterDialer struct {
	modelx.Dialer
}

// DialContext implements Dialer.DialContext
func (d EmitterDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	start := time.Now()
	conn, err := d.Dialer.DialContext(ctx, network, address)
	stop := time.Now()
	root := modelx.ContextMeasurementRootOrDefault(ctx)
	root.Handler.OnMeasurement(modelx.Measurement{
		Connect: &modelx.ConnectEvent{
			ConnID:                 safeConnID(network, conn),
			DialID:                 dialid.ContextDialID(ctx),
			DurationSinceBeginning: stop.Sub(root.Beginning),
			Error:                  err,
			Network:                network,
			RemoteAddress:          address,
			SyscallDuration:        stop.Sub(start),
			TransactionID:          transactionid.ContextTransactionID(ctx),
		},
	})
	return conn, err
}

func safeLocalAddress(conn net.Conn) (s string) {
	if conn != nil && conn.LocalAddr() != nil {
		s = conn.LocalAddr().String()
	}
	return
}

func safeConnID(network string, conn net.Conn) int64 {
	return connid.Compute(network, safeLocalAddress(conn))
}

// MeasuringDialer is a dialer that measures the underlying connection
type MeasuringDialer struct {
	modelx.Dialer
}

// DialContext implements Dialer.DialContext
func (d MeasuringDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	root := modelx.ContextMeasurementRootOrDefault(ctx)
	return MeasuringConn{
		Conn:      conn,
		Beginning: root.Beginning,
		Handler:   root.Handler,
		ID:        safeConnID(network, conn),
	}, nil
}

// MeasuringConn is a net.Conn used to perform measurements
type MeasuringConn struct {
	net.Conn
	Beginning time.Time
	Handler   modelx.Handler
	ID        int64
}

// Read reads data from the connection.
func (c MeasuringConn) Read(b []byte) (n int, err error) {
	start := time.Now()
	n, err = c.Conn.Read(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "read",
	}.MaybeBuild()
	stop := time.Now()
	c.Handler.OnMeasurement(modelx.Measurement{
		Read: &modelx.ReadEvent{
			ConnID:                 c.ID,
			DurationSinceBeginning: stop.Sub(c.Beginning),
			Error:                  err,
			NumBytes:               int64(n),
			SyscallDuration:        stop.Sub(start),
		},
	})
	return
}

// Write writes data to the connection
func (c MeasuringConn) Write(b []byte) (n int, err error) {
	start := time.Now()
	n, err = c.Conn.Write(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "write",
	}.MaybeBuild()
	stop := time.Now()
	c.Handler.OnMeasurement(modelx.Measurement{
		Write: &modelx.WriteEvent{
			ConnID:                 c.ID,
			DurationSinceBeginning: stop.Sub(c.Beginning),
			Error:                  err,
			NumBytes:               int64(n),
			SyscallDuration:        stop.Sub(start),
		},
	})
	return
}

// Close closes the connection
func (c MeasuringConn) Close() (err error) {
	start := time.Now()
	err = c.Conn.Close()
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "close",
	}.MaybeBuild()
	stop := time.Now()
	c.Handler.OnMeasurement(modelx.Measurement{
		Close: &modelx.CloseEvent{
			ConnID:                 c.ID,
			DurationSinceBeginning: stop.Sub(c.Beginning),
			Error:                  err,
			SyscallDuration:        stop.Sub(start),
		},
	})
	return
}
