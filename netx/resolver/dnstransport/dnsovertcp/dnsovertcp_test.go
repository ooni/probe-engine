package dnsovertcp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

type tlsdialer struct {
	config *tls.Config
}

func (d *tlsdialer) DialTLS(network, address string) (net.Conn, error) {
	return d.DialTLSContext(context.Background(), network, address)
}

func (d *tlsdialer) DialTLSContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	return tls.Dial(network, address, d.config)
}

func TestIntegrationSuccessTLS(t *testing.T) {
	// "Dial interprets a nil configuration as equivalent to
	// the zero configuration; see the documentation of Config
	// for the defaults."
	address := "dns.quad9.net:853"
	transport := NewTransportTLS(&tlsdialer{}, address)
	if transport.Network() != "dot" {
		t.Fatal("unexpected network")
	}
	if transport.Address() != address {
		t.Fatal("unexpected address")
	}
	if err := threeRounds(transport); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationSuccessTCP(t *testing.T) {
	address := "9.9.9.9:53"
	transport := NewTransportTCP(&net.Dialer{}, address)
	if transport.Network() != "tcp" {
		t.Fatal("unexpected network")
	}
	if transport.Address() != address {
		t.Fatal("unexpected address")
	}
	if err := threeRounds(transport); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationLookupHostError(t *testing.T) {
	transport := NewTransportTCP(&net.Dialer{}, "antani.local")
	if err := roundTrip(transport, "ooni.io."); err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationCustomTLSConfig(t *testing.T) {
	transport := NewTransportTLS(&tlsdialer{
		config: &tls.Config{
			MinVersion: tls.VersionTLS10,
		},
	}, "dns.quad9.net:853")
	if err := roundTrip(transport, "ooni.io."); err != nil {
		t.Fatal(err)
	}
}

func TestUnitRoundTripWithConnFailure(t *testing.T) {
	// fakeconn will fail in the SetDeadline, therefore we will have
	// an immediate error and we expect all errors the be alike
	transport := NewTransportTCP(&fakeconnDialer{}, "8.8.8.8:53")
	query := make([]byte, 1<<10)
	reply, err := transport.doWithConn(&fakeconn{}, query)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if reply != nil {
		t.Fatal("expected nil error here")
	}
}

func threeRounds(transport *Transport) error {
	err := roundTrip(transport, "ooni.io.")
	if err != nil {
		return err
	}
	err = roundTrip(transport, "slashdot.org.")
	if err != nil {
		return err
	}
	err = roundTrip(transport, "kernel.org.")
	if err != nil {
		return err
	}
	return nil
}

func roundTrip(transport *Transport, domain string) error {
	query := new(dns.Msg)
	query.SetQuestion(domain, dns.TypeA)
	data, err := query.Pack()
	if err != nil {
		return err
	}
	data, err = transport.RoundTrip(context.Background(), data)
	if err != nil {
		return err
	}
	return query.Unpack(data)
}

type fakeconnDialer struct {
	fakeconn fakeconn
}

func (d *fakeconnDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *fakeconnDialer) DialContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	return &d.fakeconn, nil
}

type fakeconn struct{}

func (fakeconn) Read(b []byte) (n int, err error) {
	n = len(b)
	return
}
func (fakeconn) Write(b []byte) (n int, err error) {
	n = len(b)
	return
}
func (fakeconn) Close() (err error) {
	return
}
func (fakeconn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}
func (fakeconn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}
func (fakeconn) SetDeadline(t time.Time) (err error) {
	return errors.New("cannot set deadline")
}
func (fakeconn) SetReadDeadline(t time.Time) (err error) {
	return
}
func (fakeconn) SetWriteDeadline(t time.Time) (err error) {
	return
}

func TestTLSDialerAdapter(t *testing.T) {
	fake := &fakeTLSDialer{}
	adapter := newTLSDialerAdapter(fake)
	adapter.Dial("tcp", "www.google.com:443")
	if !fake.calledDialTLS {
		t.Fatal("redirection to DialTLS not working")
	}
	adapter.DialContext(context.Background(), "tcp", "www.google.com:443")
	if !fake.calledDialTLSContext {
		t.Fatal("redirection to DialTLSContext not working")
	}
}

type fakeTLSDialer struct {
	calledDialTLS        bool
	calledDialTLSContext bool
}

func (d *fakeTLSDialer) DialTLS(network, address string) (net.Conn, error) {
	d.calledDialTLS = true
	return nil, errors.New("mocked error")
}

func (d *fakeTLSDialer) DialTLSContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	d.calledDialTLSContext = true
	return nil, errors.New("mocked error")
}
