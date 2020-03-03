package netx_test

import (
	"context"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

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

func TestHTTPIntegration(t *testing.T) {
	client := netx.NewHTTPClientWithoutProxy(handlers.NoHandler)
	defer client.Transport.CloseIdleConnections()
	err := client.ConfigureDNS("udp", "1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.HTTPClient.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPIntegrationSetResolver(t *testing.T) {
	client := netx.NewHTTPClientWithoutProxy(handlers.NoHandler)
	defer client.Transport.CloseIdleConnections()
	client.SetResolver(new(net.Resolver))
	resp, err := client.HTTPClient.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}
func TestHTTPSetCABundle(t *testing.T) {
	client := netx.NewHTTPClientWithoutProxy(handlers.NoHandler)
	err := client.SetCABundle("testdata/cacert.pem")
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPForceSpecificSNI(t *testing.T) {
	client := netx.NewHTTPClientWithoutProxy(handlers.NoHandler)
	err := client.ForceSpecificSNI("www.facebook.com")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.HTTPClient.Get("https://www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	// TODO(bassosimone): how to unwrap the error in Go < 1.13? Anyway we are
	// already testing we're getting the right error in netx_test.go.
	t.Log(err)
	if resp != nil {
		t.Fatal("expected a nil response here")
	}
}

func TestHTTPForceSkipVerify(t *testing.T) {
	client := netx.NewHTTPClientWithoutProxy(handlers.NoHandler)
	client.ForceSkipVerify()
	resp, err := client.HTTPClient.Get("https://self-signed.badssl.com/")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non nil response here")
	}
}

func TestHTTPNewClientWithoutProxy(t *testing.T) {
	client := netx.NewHTTPClientWithoutProxy(handlers.NoHandler)
	proxyTestMain(t, client.HTTPClient, 200)
}

func TestHTTPNewClientHonoursProxy(t *testing.T) {
	client := netx.NewHTTPClient(handlers.NoHandler)
	proxyTestMain(t, client.HTTPClient, 451)
}

func TestHTTPNewTransportHonoursProxy(t *testing.T) {
	transport := netx.NewHTTPTransport(
		time.Now(), handlers.NoHandler,
	)
	client := &http.Client{Transport: transport}
	proxyTestMain(t, client, 451)
}

func TestHTTPNewTransportWithoutAnyProxy(t *testing.T) {
	transport := netx.NewHTTPTransportWithProxyFunc(nil)
	client := &http.Client{Transport: transport}
	proxyTestMain(t, client, 200)
}

func proxyTestMain(t *testing.T, client *http.Client, expect int) {
	req, err := http.NewRequest("GET", "http://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != expect {
		t.Fatal("unexpected status code")
	}
}

var (
	proxyServer *httptest.Server
	proxyCount  int64
)

func TestMain(m *testing.M) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&proxyCount, 1)
			w.WriteHeader(451)
		}))
	defer server.Close()
	os.Setenv("HTTP_PROXY", server.URL)
	os.Exit(m.Run())
}
