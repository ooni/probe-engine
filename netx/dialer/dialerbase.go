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

// BaseDialer is a net.BaseDialer that is only able to connect to
// remote TCP/UDP endpoints. DNS is not supported.
type BaseDialer struct {
	dialer    modelx.Dialer
	beginning time.Time
	handler   modelx.Handler
}

// NewBaseDialer creates a new BaseDialer
func NewBaseDialer(
	beginning time.Time,
	handler modelx.Handler,
	dialer modelx.Dialer,
) *BaseDialer {
	return &BaseDialer{
		dialer:    dialer,
		beginning: beginning,
		handler:   handler,
	}
}

// Dial creates a TCP or UDP connection. See net.Dial docs.
func (d *BaseDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext dials a new connection with context.
func (d *BaseDialer) DialContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	dialID := dialid.ContextDialID(ctx)
	// this is the same timeout used by Go's net/http.DefaultTransport
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	start := time.Now()
	conn, err := d.dialer.DialContext(ctx, network, address)
	stop := time.Now()
	err = errwrapper.SafeErrWrapperBuilder{
		// ConnID does not make any sense if we've failed and the error
		// does not make any sense (and is nil) if we succeded.
		DialID:    dialID,
		Error:     err,
		Operation: "connect",
	}.MaybeBuild()
	connID := safeConnID(network, conn)
	txID := transactionid.ContextTransactionID(ctx)
	d.handler.OnMeasurement(modelx.Measurement{
		Connect: &modelx.ConnectEvent{
			ConnID:                 connID,
			DialID:                 dialID,
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
	return &MeasuringConn{
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
