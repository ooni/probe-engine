package resolver

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestIntegrationDNSOverUDPSuccessWithAddress(t *testing.T) {
	const address = "9.9.9.9:53"
	transport := NewDNSOverUDP(
		&net.Dialer{}, address,
	)
	if transport.Network() != "udp" {
		t.Fatal("invalid network")
	}
	if transport.Address() != address {
		t.Fatal("invalid address")
	}
	err := threeRounds(transport)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationDNSOverUDPSuccessWithDomain(t *testing.T) {
	transport := NewDNSOverUDP(
		&net.Dialer{}, "dns.quad9.net:53",
	)
	err := threeRounds(transport)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationDNSOverUDPDialFailure(t *testing.T) {
	transport := NewDNSOverUDP(
		&failingDialer{}, "9.9.9.9:53",
	)
	err := threeRounds(transport)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationDNSOverUDPSetDeadlineError(t *testing.T) {
	transport := NewDNSOverUDP(
		&fakeconnDialer{
			fakeconn: fakeconn{
				setDeadlineError: errors.New("mocked error"),
			},
		}, "9.9.9.9:53",
	)
	err := threeRounds(transport)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationDNSOverUDPWriteError(t *testing.T) {
	transport := NewDNSOverUDP(
		&fakeconnDialer{
			fakeconn: fakeconn{
				writeError: errors.New("mocked error"),
			},
		}, "9.9.9.9:53",
	)
	err := threeRounds(transport)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

type failingDialer struct{}

func (d *failingDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *failingDialer) DialContext(
	ctx context.Context, network, address string,
) (net.Conn, error) {
	return nil, errors.New("mocked error")
}
