// Package connx contains net.Conn extensions
package connx

import (
	"net"
	"time"

	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/modelx"
)

// MeasuringConn is a net.Conn used to perform measurements
type MeasuringConn struct {
	net.Conn
	Beginning time.Time
	Handler   modelx.Handler
	ID        int64
}

// Read reads data from the connection.
func (c *MeasuringConn) Read(b []byte) (n int, err error) {
	start := time.Now()
	n, err = c.Conn.Read(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "read",
	}.MaybeBuild()
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
func (c *MeasuringConn) Write(b []byte) (n int, err error) {
	start := time.Now()
	n, err = c.Conn.Write(b)
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "write",
	}.MaybeBuild()
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
func (c *MeasuringConn) Close() (err error) {
	start := time.Now()
	err = c.Conn.Close()
	err = errwrapper.SafeErrWrapperBuilder{
		ConnID:    c.ID,
		Error:     err,
		Operation: "close",
	}.MaybeBuild()
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
