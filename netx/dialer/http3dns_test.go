package dialer_test

import (
	"crypto/tls"
	"errors"
	"net"
	"testing"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/netx/dialer"
)

func TestHTTP3DNSDialerNoPort(t *testing.T) {
	dialer := dialer.HTTP3DNSDialer{Resolver: new(net.Resolver)}
	conn, err := dialer.Dial("udp", "antani.ooni.nu", &tls.Config{}, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestHTTP3DNSDialerLookupHostAddress(t *testing.T) {
	dialer := dialer.HTTP3DNSDialer{Resolver: MockableResolver{
		Err: errors.New("mocked error"),
	}}
	addrs, err := dialer.LookupHost("1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0] != "1.1.1.1" {
		t.Fatal("not the result we expected")
	}
}

func TestHTTP3DNSDialerLookupHostFailure(t *testing.T) {
	expected := errors.New("mocked error")
	dialer := dialer.HTTP3DNSDialer{Resolver: MockableResolver{
		Err: expected,
	}}
	conn, err := dialer.Dial("udp", "dns.google.com:853", &tls.Config{}, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected nil conn")
	}
}
