package dialer_test

import (
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/netx/dialer"
)

func TestHTTP3DNSDialerSuccess(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.HTTP3DNSDialer{Resolver: new(net.Resolver)}
	sess, err := dialer.Dial("udp", "www.google.com:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal("unexpected error")
	}
	if sess == nil {
		t.Fatal("non nil sess expected")
	}
}
func TestHTTP3DNSDialerNoPort(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.HTTP3DNSDialer{Resolver: new(net.Resolver)}
	sess, err := dialer.Dial("udp", "antani.ooni.nu", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected a nil sess here")
	}
	if err.Error() != "address antani.ooni.nu: missing port in address" {
		t.Fatal("not the error we expected")
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
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	expected := errors.New("mocked error")
	dialer := dialer.HTTP3DNSDialer{Resolver: MockableResolver{
		Err: expected,
	}}
	sess, err := dialer.Dial("udp", "dns.google.com:853", tlsConf, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
}

func TestHTTP3DNSDialerInvalidPort(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.HTTP3DNSDialer{Resolver: new(net.Resolver)}
	sess, err := dialer.Dial("udp", "www.google.com:0", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
	if !strings.HasSuffix(err.Error(), "sendto: invalid argument") &&
		!strings.HasSuffix(err.Error(), "sendto: can't assign requested address") {
		t.Fatal("not the error we expected")
	}
}

func TestHTTP3DNSDialerInvalidPortSyntax(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.HTTP3DNSDialer{Resolver: new(net.Resolver)}
	sess, err := dialer.Dial("udp", "www.google.com:port", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Fatal("not the error we expected")
	}
}

func TestHTTP3DNSDialerNilTLSConf(t *testing.T) {
	dialer := dialer.HTTP3DNSDialer{Resolver: new(net.Resolver)}
	sess, err := dialer.Dial("udp", "www.google.com:443", nil, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
	if err.Error() != "quic: tls.Config not set" {
		t.Fatal("not the error we expected")
	}
}

func TestHTTP3DNSDialerDialEarlyFails(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	expected := errors.New("mocked DialEarly error")
	mockDialEarly := func(net.PacketConn, net.Addr, string, *tls.Config, *quic.Config) (quic.EarlySession, error) {
		return nil, expected
	}
	dialer := dialer.HTTP3DNSDialer{Resolver: new(net.Resolver), DialEarly: mockDialEarly}
	sess, err := dialer.Dial("udp", "www.google.com:443", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}
