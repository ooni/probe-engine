package dialer

import (
	"context"
	"net"
	"time"

	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
)

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
