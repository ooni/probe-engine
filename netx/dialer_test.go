package netx_test

import (
	"crypto/x509"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/netx"
)

func TestIntegrationDialerDial(t *testing.T) {
	dialer := netx.NewDialer()
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialerDialWithCustomResolver(t *testing.T) {
	dialer := netx.NewDialer()
	resolver, err := netx.NewResolver("tcp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	dialer.SetResolver(resolver)
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialerDialWithConfigureDNS(t *testing.T) {
	dialer := netx.NewDialer()
	err := dialer.ConfigureDNS("tcp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := dialer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialerDialTLS(t *testing.T) {
	dialer := netx.NewDialer()
	conn, err := dialer.DialTLS("tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialerDialTLSForceSkipVerify(t *testing.T) {
	dialer := netx.NewDialer()
	dialer.ForceSkipVerify()
	conn, err := dialer.DialTLS("tcp", "self-signed.badssl.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDialerSetCABundleNonexisting(t *testing.T) {
	dialer := netx.NewDialer()
	err := dialer.SetCABundle("testdata/cacert-nonexistent.pem")
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationDialerSetCABundleInvalid(t *testing.T) {
	dialer := netx.NewDialer()
	err := dialer.SetCABundle("testdata/cacert-invalid.pem")
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationDialerSetCABundleWAI(t *testing.T) {
	dialer := netx.NewDialer()
	err := dialer.SetCABundle("testdata/cacert.pem")
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

func TestIntegrationDialerForceSpecificSNI(t *testing.T) {
	dialer := netx.NewDialer()
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
