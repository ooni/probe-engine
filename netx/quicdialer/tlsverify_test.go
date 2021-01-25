package quicdialer_test

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"regexp"
	"testing"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/netx/quicdialer"
)

type MockRedirecter struct {
	Host string
}

func (r MockRedirecter) DialContext(ctx context.Context, network, host string,
	tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	return quic.DialAddrEarly(r.Host, tlsCfg, cfg)

}

func TestTLSVerifierSuccess(t *testing.T) {
	dialer := quicdialer.TLSVerifier{
		Dialer: quicdialer.DNSDialer{Resolver: new(net.Resolver), Dialer: quicdialer.SystemDialer{}},
	}
	tlsConf := &tls.Config{NextProtos: []string{"h3-29"}, InsecureSkipVerify: true}
	sess, err := dialer.DialContext(context.Background(), "udp", "www.google.com:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal("unexpected error")
	}
	if sess == nil {
		t.Fatal("non nil sess expected")
	}
}
func TestTLSVerifierSuccessExampleSNI(t *testing.T) {
	dialer := quicdialer.TLSVerifier{
		Dialer: quicdialer.DNSDialer{Resolver: new(net.Resolver), Dialer: quicdialer.SystemDialer{}},
	}
	sni := "example.com"
	tlsConf := &tls.Config{NextProtos: []string{"h3-29"}, ServerName: sni, InsecureSkipVerify: true}
	sess, err := dialer.DialContext(context.Background(), "udp", "www.google.com:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal("unexpected error")
	}
	if sess == nil {
		t.Fatal("non nil sess expected")
	}
}
func TestTLSVerifierFailure(t *testing.T) {
	expected := errors.New("mocked error")
	dialer := quicdialer.TLSVerifier{
		Dialer: MockDialer{Err: expected},
	}
	tlsConf := &tls.Config{NextProtos: []string{"h3-29"}, InsecureSkipVerify: true}
	_, err := dialer.DialContext(context.Background(), "udp", "www.google.com:443", tlsConf, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}
func TestTLSVerifierFailureInvalidHost(t *testing.T) {
	dialer := quicdialer.TLSVerifier{
		Dialer: MockDialer{Dialer: MockRedirecter{Host: "www.google.com:443"}},
	}
	tlsConf := &tls.Config{NextProtos: []string{"h3-29"}, InsecureSkipVerify: true}
	sess, err := dialer.DialContext(context.Background(), "udp", "www.google.com", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected a nil session here")
	}
	if err.Error() != "address www.google.com: missing port in address" {
		t.Fatal("not the error we expected")
	}
}
func TestTLSVerifierInvalidCertificate(t *testing.T) {
	redirecthost := "www.cloudflare.com:443"
	dialer := quicdialer.TLSVerifier{
		Dialer: MockDialer{Dialer: MockRedirecter{Host: redirecthost}},
	}
	tlsConf := &tls.Config{NextProtos: []string{"h3-29"}, InsecureSkipVerify: true}
	sess, err := dialer.DialContext(context.Background(), "udp", "www.google.com:443", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected a nil session here")
	}
	matched, err := regexp.MatchString(`.*x509: certificate is valid for.*not.*`, err.Error())
	if !matched {
		t.Fatal("not the error we expected")
	}
}
