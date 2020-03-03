package httpx_test

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/httpx"
)

func TestIntegration(t *testing.T) {
	client := httpx.NewClientWithoutProxy(handlers.NoHandler)
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

func TestIntegrationSetResolver(t *testing.T) {
	client := httpx.NewClientWithoutProxy(handlers.NoHandler)
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
func TestSetCABundle(t *testing.T) {
	client := httpx.NewClientWithoutProxy(handlers.NoHandler)
	err := client.SetCABundle("../testdata/cacert.pem")
	if err != nil {
		t.Fatal(err)
	}
}

func TestForceSpecificSNI(t *testing.T) {
	client := httpx.NewClientWithoutProxy(handlers.NoHandler)
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

func TestForceSkipVerify(t *testing.T) {
	client := httpx.NewClientWithoutProxy(handlers.NoHandler)
	client.ForceSkipVerify()
	resp, err := client.HTTPClient.Get("https://self-signed.badssl.com/")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non nil response here")
	}
}

func TestNewClientWithoutProxy(t *testing.T) {
	client := httpx.NewClientWithoutProxy(handlers.NoHandler)
	proxyTestMain(t, client.HTTPClient, 200)
}

func TestNewClientHonoursProxy(t *testing.T) {
	client := httpx.NewClient(handlers.NoHandler)
	proxyTestMain(t, client.HTTPClient, 451)
}

func TestNewTransportHonoursProxy(t *testing.T) {
	transport := httpx.NewTransport(
		time.Now(), handlers.NoHandler,
	)
	client := &http.Client{Transport: transport}
	proxyTestMain(t, client, 451)
}

func TestNewTransportWithoutAnyProxy(t *testing.T) {
	transport := httpx.NewTransportWithProxyFunc(nil)
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
