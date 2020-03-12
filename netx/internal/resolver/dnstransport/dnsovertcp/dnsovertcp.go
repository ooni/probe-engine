// Package dnsovertcp implements DNS over TCP. It is possible to
// use both plaintext TCP and TLS.
package dnsovertcp

import (
	"bufio"
	"context"
	"io"
	"net"
	"time"

	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/netx/modelx"
)

// Transport is a DNS over TCP/TLS modelx.DNSRoundTripper.
//
// As a known bug, this implementation always creates a new connection
// for each incoming query, thus increasing the response delay.
type Transport struct {
	dialer          dialerAdapter
	address         string
	requiresPadding bool
}

type dialerAdapter interface {
	modelx.Dialer
	Network() string
}

// NewTransportTCP creates a new TCP Transport
func NewTransportTCP(dialer modelx.Dialer, address string) *Transport {
	return &Transport{
		dialer:          newTCPDialerAdapter(dialer),
		address:         address,
		requiresPadding: false,
	}
}

// NewTransportTLS creates a new TLS Transport
func NewTransportTLS(dialer modelx.TLSDialer, address string) *Transport {
	return &Transport{
		dialer:          newTLSDialerAdapter(dialer),
		address:         address,
		requiresPadding: true,
	}
}

// RoundTrip sends a request and receives a response.
func (t *Transport) RoundTrip(ctx context.Context, query []byte) ([]byte, error) {
	conn, err := t.dialer.DialContext(ctx, "tcp", t.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return t.doWithConn(conn, query)
}

// RequiresPadding returns true for DoT and false for TCP
// according to RFC8467.
func (t *Transport) RequiresPadding() bool {
	return t.requiresPadding
}

func (t *Transport) doWithConn(conn net.Conn, query []byte) (reply []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			reply = nil // we already got the error just clear the reply
		}
	}()
	err = conn.SetDeadline(time.Now().Add(10 * time.Second))
	runtimex.PanicOnError(err, "conn.SetDeadline failed")
	// Write request
	writer := bufio.NewWriter(conn)
	err = writer.WriteByte(byte(len(query) >> 8))
	runtimex.PanicOnError(err, "writer.WriteByte failed for first byte")
	err = writer.WriteByte(byte(len(query)))
	runtimex.PanicOnError(err, "writer.WriteByte failed for second byte")
	_, err = writer.Write(query)
	runtimex.PanicOnError(err, "writer.Write failed for query")
	err = writer.Flush()
	runtimex.PanicOnError(err, "writer.Flush failed")
	// Read response
	header := make([]byte, 2)
	_, err = io.ReadFull(conn, header)
	runtimex.PanicOnError(err, "io.ReadFull failed")
	length := int(header[0])<<8 | int(header[1])
	reply = make([]byte, length)
	_, err = io.ReadFull(conn, reply)
	runtimex.PanicOnError(err, "io.ReadFull failed")
	return reply, nil
}

type tlsDialerAdapter struct {
	dialer modelx.TLSDialer
}

func newTLSDialerAdapter(dialer modelx.TLSDialer) *tlsDialerAdapter {
	return &tlsDialerAdapter{dialer: dialer}
}

func (d *tlsDialerAdapter) Dial(network, address string) (net.Conn, error) {
	return d.dialer.DialTLS(network, address)
}

func (d *tlsDialerAdapter) DialContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	return d.dialer.DialTLSContext(ctx, network, address)
}

func (d *tlsDialerAdapter) Network() string {
	return "dot"
}

type tcpDialerAdapter struct {
	modelx.Dialer
}

func newTCPDialerAdapter(dialer modelx.Dialer) *tcpDialerAdapter {
	return &tcpDialerAdapter{Dialer: dialer}
}

func (d *tcpDialerAdapter) Network() string {
	return "tcp"
}

// Network returns the transport network (e.g., doh, dot)
func (t *Transport) Network() string {
	return t.dialer.Network()
}

// Address returns the upstream server address.
func (t *Transport) Address() string {
	return t.address
}
