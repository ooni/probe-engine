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

// Dialer is a dialer for network connections.
type Dialer interface {
	// DialContext is like Dial but with context
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Resolver is a DNS resolver. The *net.Resolver used by Go implements
// this interface, but other implementations are possible.
type Resolver interface {
	// LookupHost resolves a hostname to a list of IP addresses.
	LookupHost(ctx context.Context, hostname string) (addrs []string, err error)
}

// TimeoutDialer is a wrapper for the system dialer
type TimeoutDialer struct {
	Dialer
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
	Dialer
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
	if err != nil {
		return nil, err
	}
	return &ErrWrapperConn{Conn: conn, ID: safeConnID(network, conn)}, err
}

// ErrWrapperConn is a net.Conn that performs error wrapping.
type ErrWrapperConn struct {
	net.Conn
	ID int64
}

// Read reads data from the connection.
func (c ErrWrapperConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "read",
	}.MaybeBuild()
	return
}

// Write writes data to the connection
func (c ErrWrapperConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "write",
	}.MaybeBuild()
	return
}

// Close closes the connection
func (c ErrWrapperConn) Close() (err error) {
	err = c.Conn.Close()
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "close",
	}.MaybeBuild()
	return
}

// EmitterDialer is a dialer that emits events
type EmitterDialer struct {
	Dialer
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
	if err != nil {
		return nil, err
	}
	return EmitterConn{
		Conn:      conn,
		Beginning: root.Beginning,
		Handler:   root.Handler,
		ID:        safeConnID(network, conn),
	}, nil
}

// EmitterConn is a net.Conn used to emit measurements
type EmitterConn struct {
	net.Conn
	Beginning time.Time
	Handler   modelx.Handler
	ID        int64
}

// Read reads data from the connection.
func (c EmitterConn) Read(b []byte) (n int, err error) {
	start := time.Now()
	n, err = c.Conn.Read(b)
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
func (c EmitterConn) Write(b []byte) (n int, err error) {
	start := time.Now()
	n, err = c.Conn.Write(b)
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
func (c EmitterConn) Close() (err error) {
	start := time.Now()
	err = c.Conn.Close()
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

func safeLocalAddress(conn net.Conn) (s string) {
	if conn != nil && conn.LocalAddr() != nil {
		s = conn.LocalAddr().String()
	}
	return
}

func safeConnID(network string, conn net.Conn) int64 {
	return connid.Compute(network, safeLocalAddress(conn))
}
