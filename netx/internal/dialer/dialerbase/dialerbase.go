// Package dialerbase contains the base dialer functionality. We connect
// to a remote endpoint, but we don't support DNS.
package dialerbase

import (
	"context"
	"net"
	"time"

	"github.com/ooni/probe-engine/netx/internal/connid"
	"github.com/ooni/probe-engine/netx/internal/dialer/connx"
	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// Dialer is a net.Dialer that is only able to connect to
// remote TCP/UDP endpoints. DNS is not supported.
type Dialer struct {
	dialer    modelx.Dialer
	beginning time.Time
	handler   modelx.Handler
	dialID    int64
}

// New creates a new dialer
func New(
	beginning time.Time,
	handler modelx.Handler,
	dialer modelx.Dialer,
	dialID int64,
) *Dialer {
	return &Dialer{
		dialer:    dialer,
		beginning: beginning,
		handler:   handler,
		dialID:    dialID,
	}
}

// Dial creates a TCP or UDP connection. See net.Dial docs.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext dials a new connection with context.
func (d *Dialer) DialContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	// this is the same timeout used by Go's net/http.DefaultTransport
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	start := time.Now()
	conn, err := d.dialer.DialContext(ctx, network, address)
	stop := time.Now()
	err = errwrapper.SafeErrWrapperBuilder{
		// ConnID does not make any sense if we've failed and the error
		// does not make any sense (and is nil) if we succeded.
		DialID:    d.dialID,
		Error:     err,
		Operation: "connect",
	}.MaybeBuild()
	connID := safeConnID(network, conn)
	txID := transactionid.ContextTransactionID(ctx)
	d.handler.OnMeasurement(modelx.Measurement{
		Connect: &modelx.ConnectEvent{
			ConnID:                 connID,
			DialID:                 d.dialID,
			DurationSinceBeginning: stop.Sub(d.beginning),
			Error:                  err,
			Network:                network,
			RemoteAddress:          address,
			SyscallDuration:        stop.Sub(start),
			TransactionID:          txID,
		},
	})
	if err != nil {
		return nil, err
	}
	return &connx.MeasuringConn{
		Conn:      conn,
		Beginning: d.beginning,
		Handler:   d.handler,
		ID:        connID,
	}, nil
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
