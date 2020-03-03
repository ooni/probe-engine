package internal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/internal/resolver/brokenresolver"
	"github.com/ooni/probe-engine/netx/internal/resolver/systemresolver"
)

func TestIntegrationDial(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialWithCustomResolver(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	dialer.SetResolver(new(net.Resolver))
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialTLS(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialTLSForceSkipVerify(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	dialer.ForceSkipVerify()
	conn, err := dialer.DialTLS("tcp", "self-signed.badssl.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialInvalidAddress(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	conn, err := dialer.Dial("tcp", "www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestIntegrationDialInvalidAddressTLS(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	conn, err := dialer.DialTLS("tcp", "www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestIntegrationDialInvalidSNI(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	dialer.TLSConfig = &tls.Config{
		ServerName: "www.google.com",
	}
	conn, err := dialer.DialTLS("tcp", "ooni.io:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestDialerSetCABundleExisting(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	err := dialer.SetCABundle("../testdata/cacert.pem")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDialerSetCABundleNonexisting(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	err := dialer.SetCABundle("../testdata/cacert-nonexistent.pem")
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestDialerSetCABundleWAI(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	err := dialer.SetCABundle("../testdata/cacert.pem")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	var target x509.UnknownAuthorityError
	if errors.As(err, &target) == false {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestDialerForceSpecificSNI(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	err := dialer.ForceSpecificSNI("www.facebook.com")
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	var target x509.HostnameError
	if errors.As(err, &target) == false {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected a nil connection here")
	}
}

func testresolverquick(t *testing.T, network, address string) {
	resolver, err := NewResolver(time.Now(), handlers.NoHandler, network, address)
	if err != nil {
		t.Fatal(err)
	}
	if resolver == nil {
		t.Fatal("expected non-nil resolver here")
	}
	addrs, err := resolver.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expected non-nil addrs here")
	}
	var foundquad8 bool
	for _, addr := range addrs {
		if addr == "8.8.8.8" {
			foundquad8 = true
		}
	}
	if !foundquad8 {
		t.Fatal("did not find 8.8.8.8 in ouput")
	}
}

func TestIntegrationNewResolverUDPAddress(t *testing.T) {
	testresolverquick(t, "udp", "8.8.8.8:53")
}

func TestIntegrationNewResolverUDPAddressNoPort(t *testing.T) {
	testresolverquick(t, "udp", "8.8.8.8")
}

func TestIntegrationNewResolverUDPDomain(t *testing.T) {
	testresolverquick(t, "udp", "dns.google.com:53")
}

func TestIntegrationNewResolverUDPDomainNoPort(t *testing.T) {
	testresolverquick(t, "udp", "dns.google.com")
}

func TestIntegrationNewResolverSystem(t *testing.T) {
	testresolverquick(t, "system", "")
}

func TestIntegrationNewResolverTCPAddress(t *testing.T) {
	testresolverquick(t, "tcp", "8.8.8.8:53")
}

func TestIntegrationNewResolverTCPAddressNoPort(t *testing.T) {
	testresolverquick(t, "tcp", "8.8.8.8")
}

func TestIntegrationNewResolverTCPDomain(t *testing.T) {
	testresolverquick(t, "tcp", "dns.google.com:53")
}

func TestIntegrationNewResolverTCPDomainNoPort(t *testing.T) {
	testresolverquick(t, "tcp", "dns.google.com")
}

func TestIntegrationNewResolverDoTAddress(t *testing.T) {
	testresolverquick(t, "dot", "9.9.9.9:853")
}

func TestIntegrationNewResolverDoTAddressNoPort(t *testing.T) {
	testresolverquick(t, "dot", "9.9.9.9")
}

func TestIntegrationNewResolverDoTDomain(t *testing.T) {
	testresolverquick(t, "dot", "dns.quad9.net:853")
}

func TestIntegrationNewResolverDoTDomainNoPort(t *testing.T) {
	testresolverquick(t, "dot", "dns.quad9.net")
}

func TestIntegrationNewResolverDoH(t *testing.T) {
	testresolverquick(t, "doh", "https://cloudflare-dns.com/dns-query")
}

func TestIntegrationNewResolverInvalid(t *testing.T) {
	resolver, err := NewResolver(
		time.Now(), handlers.StdoutHandler,
		"antani", "https://cloudflare-dns.com/dns-query",
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resolver != nil {
		t.Fatal("expected a nil resolver here")
	}
}

func testconfigurednsquick(t *testing.T, network, address string) {
	d := NewDialer(time.Now(), handlers.NoHandler)
	err := d.ConfigureDNS(network, address)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := d.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("expected non-nil conn here")
	}
	conn.Close()
}

func TestIntegrationConfigureSystemDNS(t *testing.T) {
	testconfigurednsquick(t, "system", "")
}

func TestIntegrationHTTPTransport(t *testing.T) {
	client := &http.Client{
		Transport: NewHTTPTransport(
			time.Now(), handlers.NoHandler,
			NewDialer(time.Now(), handlers.NoHandler),
			false,
			http.ProxyFromEnvironment,
		),
	}
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	client.CloseIdleConnections()
}

func TestIntegrationHTTPTransportTimeout(t *testing.T) {
	client := &http.Client{
		Transport: NewHTTPTransport(
			time.Now(), handlers.NoHandler,
			NewDialer(time.Now(), handlers.NoHandler),
			false,
			http.ProxyFromEnvironment,
		),
	}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "generic_timeout_error") {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil resp here")
	}
}
func TestIntegrationChainResolvers(t *testing.T) {
	dialer := NewDialer(time.Now(), handlers.NoHandler)
	resolver := ChainResolvers(
		brokenresolver.New(),
		systemresolver.New(new(net.Resolver)),
	)
	dialer.SetResolver(resolver)
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}

func TestIntegrationFailure(t *testing.T) {
	client := &http.Client{
		Transport: NewHTTPTransport(
			time.Now(), handlers.NoHandler,
			NewDialer(time.Now(), handlers.NoHandler),
			false,
			http.ProxyFromEnvironment,
		),
	}
	// This fails the request because we attempt to speak cleartext HTTP with
	// a server that instead is expecting TLS.
	resp, err := client.Get("http://www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resp != nil {
		t.Fatal("expected a nil response here")
	}
	client.CloseIdleConnections()
}

func TestLookupAddrWrapper(t *testing.T) {
	resolver := newResolverWrapper(time.Now(), handlers.NoHandler, new(net.Resolver))
	names, err := resolver.LookupAddr(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 1 {
		t.Fatal("unexpected result")
	}
}

func TestLookupCNAMEWrapper(t *testing.T) {
	resolver := newResolverWrapper(time.Now(), handlers.NoHandler, new(net.Resolver))
	name, err := resolver.LookupCNAME(context.Background(), "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if name == "" {
		t.Fatal("unexpected result")
	}
}

func TestLookupMXWrapper(t *testing.T) {
	resolver := newResolverWrapper(time.Now(), handlers.NoHandler, new(net.Resolver))
	entries, err := resolver.LookupMX(context.Background(), "google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 1 {
		t.Fatal("unexpected result")
	}
}

func TestLookupNSWrapper(t *testing.T) {
	resolver := newResolverWrapper(time.Now(), handlers.NoHandler, new(net.Resolver))
	entries, err := resolver.LookupNS(context.Background(), "google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 1 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNewHTTPClientForDoH(t *testing.T) {
	first := newHTTPClientForDoH(
		time.Now(), handlers.NoHandler,
	)
	second := newHTTPClientForDoH(
		time.Now(), handlers.NoHandler,
	)
	if first != second {
		t.Fatal("expected to see same client here")
	}
	third := newHTTPClientForDoH(
		time.Now(), handlers.StdoutHandler,
	)
	if first == third {
		t.Fatal("expected to see different client here")
	}
}
