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

// TLSDialer is the TLS dialer
type TLSDialer struct {
	TLSHandshakeTimeout time.Duration // default: 10 second
	config              *tls.Config
	dialer              modelx.Dialer
	setDeadline         func(net.Conn, time.Time) error
}

// NewTLSDialer creates a new TLSDialer
func NewTLSDialer(dialer modelx.Dialer, config *tls.Config) *TLSDialer {
	return &TLSDialer{
		TLSHandshakeTimeout: 10 * time.Second,
		config:              config,
		dialer:              dialer,
		setDeadline: func(conn net.Conn, t time.Time) error {
			return conn.SetDeadline(t)
		},
	}
}

// DialTLS dials a new TLS connection
func (d *TLSDialer) DialTLS(network, address string) (net.Conn, error) {
	return d.DialTLSContext(context.Background(), network, address)
}

// DialTLSContext is like DialTLS, but with context
func (d *TLSDialer) DialTLSContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	conn, err := d.dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	config := d.config.Clone() // avoid polluting original config
	if config.ServerName == "" {
		config.ServerName = host
	}
	err = d.setDeadline(conn, time.Now().Add(d.TLSHandshakeTimeout))
	if err != nil {
		conn.Close()
		return nil, err
	}
	connID := connid.Compute(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	tlsconn := tls.Client(conn, config)
	root := modelx.ContextMeasurementRootOrDefault(ctx)
	// Implementation note: when DialTLS is not set, the code in
	// net/http will perform the handshake. Otherwise, if DialTLS
	// is set, we will end up here. This code is still used when
	// performing non-HTTP TLS-enabled dial operations.
	root.Handler.OnMeasurement(modelx.Measurement{
		TLSHandshakeStart: &modelx.TLSHandshakeStartEvent{
			ConnID:                 connID,
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
			SNI:                    config.ServerName,
		},
	})
	err = tlsconn.Handshake()
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    connID,
		Error:     err,
		Operation: "tls_handshake",
	}.MaybeBuild()
	root.Handler.OnMeasurement(modelx.Measurement{
		TLSHandshakeDone: &modelx.TLSHandshakeDoneEvent{
			ConnID:                 connID,
			ConnectionState:        modelx.NewTLSConnectionState(tlsconn.ConnectionState()),
			Error:                  err,
			DurationSinceBeginning: time.Now().Sub(root.Beginning),
		},
	})
	conn.SetDeadline(time.Time{}) // clear deadline
	if err != nil {
		conn.Close()
		return nil, err
	}
	return tlsconn, err
}
