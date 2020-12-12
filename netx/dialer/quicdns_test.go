package dialer_test

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/netx/dialer"
)

func TestQUICDNSDialerSuccess(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.QUICDNSDialer{Resolver: new(net.Resolver), Dialer: dialer.QUICSystemDialer{}}
	sess, err := dialer.DialContext(context.Background(), "udp", "", "www.google.com:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal("unexpected error")
	}
	if sess == nil {
		t.Fatal("non nil sess expected")
	}
}

func TestQUICDNSDialerNoPort(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.QUICDNSDialer{Resolver: new(net.Resolver), Dialer: dialer.QUICSystemDialer{}}
	sess, err := dialer.DialContext(context.Background(), "udp", "", "antani.ooni.nu", tlsConf, &quic.Config{})
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

func TestQUICDNSDialerLookupHostAddress(t *testing.T) {
	dialer := dialer.QUICDNSDialer{Resolver: MockableResolver{
		Err: errors.New("mocked error"),
	}}
	addrs, err := dialer.LookupHost(context.Background(), "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0] != "1.1.1.1" {
		t.Fatal("not the result we expected")
	}
}

func TestQUICDNSDialerLookupHostFailure(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	expected := errors.New("mocked error")
	dialer := dialer.QUICDNSDialer{Resolver: MockableResolver{
		Err: expected,
	}}
	sess, err := dialer.DialContext(context.Background(), "udp", "", "dns.google.com:853", tlsConf, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
}

func TestQUICDNSDialerInvalidPort(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.QUICDNSDialer{Resolver: new(net.Resolver), Dialer: dialer.QUICSystemDialer{}}
	sess, err := dialer.DialContext(context.Background(), "udp", "", "www.google.com:0", tlsConf, &quic.Config{})
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

func TestQUICDNSDialerInvalidPortSyntax(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	dialer := dialer.QUICDNSDialer{Resolver: new(net.Resolver), Dialer: dialer.QUICSystemDialer{}}
	sess, err := dialer.DialContext(context.Background(), "udp", "", "www.google.com:port", tlsConf, &quic.Config{})
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

func TestQUICDNSDialerNilTLSConf(t *testing.T) {
	dialer := dialer.QUICDNSDialer{Resolver: new(net.Resolver), Dialer: dialer.QUICSystemDialer{}}
	sess, err := dialer.DialContext(context.Background(), "udp", "", "www.google.com:443", nil, &quic.Config{})
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

type MockDialer struct {
	err error
}

func (d MockDialer) DialContext(context.Context, string, string, string, *tls.Config, *quic.Config) (quic.EarlySession, error) {
	return nil, d.err
}

func TestQUICDNSDialerDialEarlyFails(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	expected := errors.New("mocked DialEarly error")

	dialer := dialer.QUICDNSDialer{Resolver: new(net.Resolver), Dialer: MockDialer{expected}}
	sess, err := dialer.DialContext(context.Background(), "udp", "", "www.google.com:443", tlsConf, &quic.Config{})
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
