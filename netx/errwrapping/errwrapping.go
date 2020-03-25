// Package errwrapping contains code to wrap errors.
package errwrapping

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/measurable"
	"github.com/ooni/probe-engine/netx/modelx"
)

// ErrWrapper adds error wrapping to measurable operations
type ErrWrapper struct {
	measurable.Operations
}

// LookupHost performs an host lookupt
func (ew ErrWrapper) LookupHost(ctx context.Context, domain string) ([]string, error) {
	addrs, err := ew.Operations.LookupHost(ctx, domain)
	return addrs, errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: "resolve",
	}.MaybeBuild()
}

// DialContext establishes a new connection
func (ew ErrWrapper) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := ew.Operations.DialContext(ctx, network, address)
	return conn, errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: "connect",
	}.MaybeBuild()
}

// Handshake performs a TLS handshake
func (ew ErrWrapper) Handshake(
	ctx context.Context, conn net.Conn, config *tls.Config) (
	net.Conn, tls.ConnectionState, error) {
	tlsconn, state, err := ew.Operations.Handshake(ctx, conn, config)
	return tlsconn, state, errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: "tls_handshake",
	}.MaybeBuild()
}

// RoundTrip performs an HTTP round trip
func (ew ErrWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := ew.Operations.RoundTrip(req)
	var errConnect measurable.ErrConnect
	op := "http_round_trip"
	if errors.Is(err, &errConnect) {
		op = "connect"
	}
	return resp, errwrapper.SafeErrWrapperBuilder{
		Error:     err,
		Operation: op,
	}.MaybeBuild()
}

// Failure returns the OONI failure string. In particular:
//
// 1. if there is no error, it returns ""
// 2. if the error is not a wrapped error, it returns unknown_failure
// 3. otherwise, it returns the OONI failure string
//
// See the netx/DESIGN.md for more info.
func Failure(err error) string {
	if err == nil {
		return ""
	}
	var wrapper *modelx.ErrWrapper
	if errors.As(err, &wrapper) == false {
		return fmt.Sprintf("unknown_failure: %s", err.Error())
	}
	return wrapper.Failure
}

// Operation returns the operation that failed. In particular:
//
// 1. if there is no error, it returns ""
// 2. if the error is not a wrapped error, it returns ""
// 3. otherwise it returns the failed operation
//
// See the netx/DESIGN.md for more info.
func Operation(err error) string {
	if err == nil {
		return ""
	}
	var wrapper *modelx.ErrWrapper
	if errors.As(err, &wrapper) == false {
		return ""
	}
	return wrapper.Operation
}
