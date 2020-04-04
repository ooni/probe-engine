package dialer

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/errwrapper"
)

// ErrorWrapperDialer is a dialer that performs err wrapping
type ErrorWrapperDialer struct {
	Dialer
}

// DialContext implements Dialer.DialContext
func (d ErrorWrapperDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
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
	return &ErrorWrapperConn{
		Conn: conn, ConnID: safeConnID(network, conn), DialID: dialID}, nil
}

// ErrorWrapperConn is a net.Conn that performs error wrapping.
type ErrorWrapperConn struct {
	net.Conn
	ConnID int64
	DialID int64
}

// Read reads data from the connection.
func (c ErrorWrapperConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ConnID,
		DialID:    c.DialID,
		Error:     err,
		Operation: "read",
	}.MaybeBuild()
	return
}

// Write writes data to the connection
func (c ErrorWrapperConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ConnID,
		DialID:    c.DialID,
		Error:     err,
		Operation: "write",
	}.MaybeBuild()
	return
}

// Close closes the connection
func (c ErrorWrapperConn) Close() (err error) {
	err = c.Conn.Close()
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ConnID,
		DialID:    c.DialID,
		Error:     err,
		Operation: "close",
	}.MaybeBuild()
	return
}
