package dslx

//
// QUIC measurements
//

import (
	"context"
	"crypto/tls"
	"io"
	"time"

	"github.com/ooni/probe-engine/pkg/logx"
	"github.com/quic-go/quic-go"
)

// QUICHandshake returns a function performing QUIC handshakes.
func QUICHandshake(rt Runtime, options ...TLSHandshakeOption) Func[*Endpoint, *QUICConnection] {
	return Operation[*Endpoint, *QUICConnection](func(ctx context.Context, input *Endpoint) (*QUICConnection, error) {
		// create trace
		trace := rt.NewTrace(rt.IDGenerator().Add(1), rt.ZeroTime(), input.Tags...)

		// create a suitable TLS configuration
		config := tlsNewConfig(input.Address, []string{"h3"}, input.Domain, rt.Logger(), options...)

		// start the operation logger
		ol := logx.NewOperationLogger(
			rt.Logger(),
			"[#%d] QUICHandshake with %s SNI=%s ALPN=%v",
			trace.Index(),
			input.Address,
			config.ServerName,
			config.NextProtos,
		)

		// setup
		udpListener := trace.NewUDPListener()
		quicDialer := trace.NewQUICDialerWithoutResolver(udpListener, rt.Logger())
		const timeout = 10 * time.Second
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// handshake
		quicConn, err := quicDialer.DialContext(ctx, input.Address, config, &quic.Config{})

		var closerConn io.Closer
		var tlsState tls.ConnectionState
		if quicConn != nil {
			closerConn = &quicCloserConn{quicConn}
			tlsState = quicConn.ConnectionState().TLS // only quicConn can be nil

			// TODO(https://github.com/ooni/probe/issues/2670).
			//
			// Start measuring for throttling here.
		}

		// possibly track established conn for late close
		rt.MaybeTrackConn(closerConn)

		// stop the operation logger
		ol.Stop(err)

		// save the observations
		rt.SaveObservations(maybeTraceToObservations(trace)...)

		// handle error case
		if err != nil {
			return nil, err
		}

		// handle success
		state := &QUICConnection{
			Address:   input.Address,
			QUICConn:  quicConn,
			Domain:    input.Domain,
			Network:   input.Network,
			TLSConfig: config,
			TLSState:  tlsState,
			Trace:     trace,
		}
		return state, nil
	})
}

// QUICConnection is an established QUIC connection. If you initialize
// manually, init at least the ones marked as MANDATORY.
type QUICConnection struct {
	// Address is the MANDATORY address we tried to connect to.
	Address string

	// QUICConn is the established QUIC conn.
	QUICConn quic.EarlyConnection

	// Domain is the OPTIONAL domain we resolved.
	Domain string

	// Network is the MANDATORY network we tried to use when connecting.
	Network string

	// TLSConfig is the config we used to establish a QUIC connection and will
	// be needed when constructing an HTTP/3 transport.
	TLSConfig *tls.Config

	// TLSState is the possibly-empty TLS connection state.
	TLSState tls.ConnectionState

	// Trace is the MANDATORY trace we're using.
	Trace Trace
}

type quicCloserConn struct {
	quic.EarlyConnection
}

func (c *quicCloserConn) Close() error {
	return c.CloseWithError(0, "")
}
