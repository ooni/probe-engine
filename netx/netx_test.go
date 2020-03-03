package netx_test

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"testing"

	"github.com/ooni/probe-engine/netx"
	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/internal/resolver/brokenresolver"
)

func TestIntegrationDialer(t *testing.T) {
	dialer := netx.NewDialerWithoutHandler()
	err := dialer.ConfigureDNS("udp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	conn, err = dialer.DialContext(
		context.Background(), "tcp", "www.google.com:80",
	)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	conn, err = dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialerWithSetResolver(t *testing.T) {
	dialer := netx.NewDialer(handlers.NoHandler)
	dialer.SetResolver(new(net.Resolver))
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	conn, err = dialer.DialContext(
		context.Background(), "tcp", "www.google.com:80",
	)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	conn, err = dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationResolver(t *testing.T) {
	dialer := netx.NewDialer(handlers.NoHandler)
	resolver, err := dialer.NewResolver("tcp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	addrs, err := resolver.LookupHost(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) < 1 {
		t.Fatal("No addresses returned")
	}
}

func TestIntegrationStandaloneResolver(t *testing.T) {
	resolver, err := netx.NewResolver(handlers.NoHandler, "tcp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	addrs, err := resolver.LookupHost(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) < 1 {
		t.Fatal("No addresses returned")
	}
}

func TestIntegrationStandaloneResolverWithoutHandler(t *testing.T) {
	resolver, err := netx.NewResolverWithoutHandler("tcp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	addrs, err := resolver.LookupHost(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) < 1 {
		t.Fatal("No addresses returned")
	}
}

func TestSetCABundle(t *testing.T) {
	dialer := netx.NewDialer(handlers.NoHandler)
	err := dialer.SetCABundle("testdata/cacert.pem")
	if err != nil {
		t.Fatal(err)
	}
}

func TestForceSpecificSNI(t *testing.T) {
	dialer := netx.NewDialer(handlers.NoHandler)
	err := dialer.ForceSpecificSNI("www.facebook.com")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	var target x509.HostnameError
	if errors.As(err, &target) == false {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

func TestIntegrationDialTLSForceSkipVerify(t *testing.T) {
	dialer := netx.NewDialer(handlers.NoHandler)
	dialer.ForceSkipVerify()
	conn, err := dialer.DialTLS("tcp", "self-signed.badssl.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestChainResolvers(t *testing.T) {
	fallback, err := netx.NewResolver(handlers.NoHandler, "udp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	dialer := netx.NewDialer(handlers.NoHandler)
	resolver := netx.ChainResolvers(brokenresolver.New(), fallback)
	dialer.SetResolver(resolver)
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}
