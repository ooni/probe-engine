package dialer

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/ooni/probe-engine/netx/internal/connid"
	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/modelx"
)

// TLSHandshaker is the generic TLS handshaker
type TLSHandshaker interface {
	Handshake(ctx context.Context, conn net.Conn, config *tls.Config) (
		net.Conn, tls.ConnectionState, error)
}

// TLSHandshakerSystem is the system TLS handshaker.
type TLSHandshakerSystem struct {
	HandshakeTimeout time.Duration // default: 10 second
}

// Handshake implements Handshaker.Handshake
func (h TLSHandshakerSystem) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	timeout := h.HandshakeTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tlsconn := tls.Client(conn, config)
	errch := make(chan error, 1)
	go func() {
		errch <- tlsconn.Handshake()
	}()
	select {
	case err := <-errch:
		if err != nil {
			return nil, tls.ConnectionState{}, err
		}
		return tlsconn, tlsconn.ConnectionState(), nil
	case <-ctx.Done():
		return nil, tls.ConnectionState{}, ctx.Err()
	}
}

// TLSHandshakerErrWrapper wraps the returned error to be an OONI error
type TLSHandshakerErrWrapper struct {
	TLSHandshaker
}

// Handshake implements Handshaker.Handshake
func (h TLSHandshakerErrWrapper) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	connID := connid.Compute(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	tlsconn, state, err := h.TLSHandshaker.Handshake(ctx, conn, config)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    connID,
		Error:     err,
		Operation: "tls_handshake",
	}.MaybeBuild()
	return tlsconn, state, err
}

// TLSHandshakerEmitter emits events using the MeasurementRoot
type TLSHandshakerEmitter struct {
	TLSHandshaker
}

// Handshake implements Handshaker.Handshake
func (h TLSHandshakerEmitter) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	connID := connid.Compute(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	root := modelx.ContextMeasurementRootOrDefault(ctx)
	root.Handler.OnMeasurement(modelx.Measurement{
		TLSHandshakeStart: &modelx.TLSHandshakeStartEvent{
			ConnID:                 connID,
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
			SNI:                    config.ServerName,
		},
	})
	tlsconn, state, err := h.TLSHandshaker.Handshake(ctx, conn, config)
	root.Handler.OnMeasurement(modelx.Measurement{
		TLSHandshakeDone: &modelx.TLSHandshakeDoneEvent{
			ConnID:                 connID,
			ConnectionState:        modelx.NewTLSConnectionState(state),
			Error:                  err,
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
		},
	})
	return tlsconn, state, err
}

// TLSDialer is the TLS dialer
type TLSDialer struct {
	Config        *tls.Config
	Dialer        Dialer
	TLSHandshaker TLSHandshaker
}

// NewTLSDialer creates a new TLSDialer
func NewTLSDialer(dialer Dialer, config *tls.Config) TLSDialer {
	return TLSDialer{
		Config: config,
		Dialer: dialer,
		TLSHandshaker: TLSHandshakerEmitter{
			TLSHandshaker: TLSHandshakerErrWrapper{
				TLSHandshaker: TLSHandshakerSystem{},
			},
		},
	}
}

// DialTLSContext is like DialTLS, but with context
func (d TLSDialer) DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Implementation note: when DialTLS is not set, the code in
	// net/http will perform the handshake. Otherwise, if DialTLS
	// is set, we will end up here. This code is still used when
	// performing non-HTTP TLS-enabled dial operations.
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	config := d.Config
	if config == nil {
		config = new(tls.Config)
	} else {
		config = config.Clone()
	}
	if config.ServerName == "" {
		config.ServerName = host
	}
	tlsconn, _, err := d.TLSHandshaker.Handshake(ctx, conn, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return tlsconn, nil
}
