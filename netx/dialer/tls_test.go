package dialer_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestUnitSystemTLSHandshakerContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immeditely cancel
	h := dialer.SystemTLSHandshaker{}
	conn, _, err := h.Handshake(ctx, dialer.EOFConn{}, new(tls.Config))
	if err != context.Canceled {
		t.Fatal("not the error that we expected")
	}
	if conn != nil {
		t.Fatal("expected nil con here")
	}
}

func TestUnitSystemTLSHandshakerEOFError(t *testing.T) {
	h := dialer.SystemTLSHandshaker{}
	conn, _, err := h.Handshake(context.Background(), dialer.EOFConn{}, &tls.Config{
		ServerName: "x.org",
	})
	if err != io.EOF {
		t.Fatal("not the error that we expected")
	}
	if conn != nil {
		t.Fatal("expected nil con here")
	}
}

func TestUnitTimeoutTLSHandshaker(t *testing.T) {
	h := dialer.TimeoutTLSHandshaker{
		TLSHandshaker:    SlowTLSHandshaker{},
		HandshakeTimeout: 200 * time.Millisecond,
	}
	conn, _, err := h.Handshake(
		context.Background(), dialer.EOFConn{}, new(tls.Config))
	if err != context.DeadlineExceeded {
		t.Fatal("not the error that we expected")
	}
	if conn != nil {
		t.Fatal("expected nil con here")
	}
}

type SlowTLSHandshaker struct{}

func (SlowTLSHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	select {
	case <-ctx.Done():
		return nil, tls.ConnectionState{}, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, tls.ConnectionState{}, io.EOF
	}
}

func TestUnitErrorWrapperTLSHandshakerFailure(t *testing.T) {
	h := dialer.ErrorWrapperTLSHandshaker{TLSHandshaker: dialer.EOFTLSHandshaker{}}
	conn, _, err := h.Handshake(
		context.Background(), dialer.EOFConn{}, new(tls.Config))
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error that we expected")
	}
	if conn != nil {
		t.Fatal("expected nil con here")
	}
	var errWrapper *modelx.ErrWrapper
	if !errors.As(err, &errWrapper) {
		t.Fatal("cannot cast to ErrWrapper")
	}
	if errWrapper.ConnID == 0 {
		t.Fatal("unexpected ConnID")
	}
	if errWrapper.Failure != modelx.FailureEOFError {
		t.Fatal("unexpected Failure")
	}
	if errWrapper.Operation != "tls_handshake" {
		t.Fatal("unexpected Operation")
	}
}

func TestUnitEmitterTLSHandshakerFailure(t *testing.T) {
	saver := &handlers.SavingHandler{}
	ctx := modelx.WithMeasurementRoot(context.Background(), &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler:   saver,
	})
	h := dialer.EmitterTLSHandshaker{TLSHandshaker: dialer.EOFTLSHandshaker{}}
	conn, _, err := h.Handshake(ctx, dialer.EOFConn{}, &tls.Config{
		ServerName: "www.kernel.org",
	})
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error that we expected")
	}
	if conn != nil {
		t.Fatal("expected nil con here")
	}
	events := saver.Read()
	if len(events) != 2 {
		t.Fatal("Wrong number of events")
	}
	if events[0].TLSHandshakeStart == nil {
		t.Fatal("missing TLSHandshakeStart event")
	}
	if events[0].TLSHandshakeStart.ConnID == 0 {
		t.Fatal("expected nonzero ConnID")
	}
	if events[0].TLSHandshakeStart.DurationSinceBeginning == 0 {
		t.Fatal("expected nonzero DurationSinceBeginning")
	}
	if events[0].TLSHandshakeStart.SNI != "www.kernel.org" {
		t.Fatal("expected nonzero SNI")
	}
	if events[1].TLSHandshakeDone == nil {
		t.Fatal("missing TLSHandshakeDone event")
	}
	if events[1].TLSHandshakeDone.ConnID == 0 {
		t.Fatal("expected nonzero ConnID")
	}
	if events[1].TLSHandshakeDone.DurationSinceBeginning == 0 {
		t.Fatal("expected nonzero DurationSinceBeginning")
	}
}

func TestUnitTLSDialerFailureSplitHostPort(t *testing.T) {
	dialer := dialer.TLSDialer{}
	conn, err := dialer.DialTLSContext(
		context.Background(), "tcp", "www.google.com") // missing port
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestUnitTLSDialerFailureDialing(t *testing.T) {
	dialer := dialer.TLSDialer{Dialer: dialer.EOFDialer{}}
	conn, err := dialer.DialTLSContext(
		context.Background(), "tcp", "www.google.com:443")
	if !errors.Is(err, io.EOF) {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
}

func TestUnitTLSDialerFailureHandshaking(t *testing.T) {
	rec := &RecorderTLSHandshaker{TLSHandshaker: dialer.SystemTLSHandshaker{}}
	dialer := dialer.TLSDialer{
		Dialer:        dialer.EOFConnDialer{},
		TLSHandshaker: rec,
	}
	conn, err := dialer.DialTLSContext(
		context.Background(), "tcp", "www.google.com:443")
	if !errors.Is(err, io.EOF) {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
	if rec.SNI != "www.google.com" {
		t.Fatal("unexpected SNI value")
	}
}

func TestUnitTLSDialerFailureHandshakingOverrideSNI(t *testing.T) {
	rec := &RecorderTLSHandshaker{TLSHandshaker: dialer.SystemTLSHandshaker{}}
	dialer := dialer.TLSDialer{
		Config: &tls.Config{
			ServerName: "x.org",
		},
		Dialer:        dialer.EOFConnDialer{},
		TLSHandshaker: rec,
	}
	conn, err := dialer.DialTLSContext(
		context.Background(), "tcp", "www.google.com:443")
	if !errors.Is(err, io.EOF) {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("connection is not nil")
	}
	if rec.SNI != "x.org" {
		t.Fatal("unexpected SNI value")
	}
}

type RecorderTLSHandshaker struct {
	dialer.TLSHandshaker
	SNI string
}

func (h *RecorderTLSHandshaker) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config,
) (net.Conn, tls.ConnectionState, error) {
	h.SNI = config.ServerName
	return h.TLSHandshaker.Handshake(ctx, conn, config)
}
