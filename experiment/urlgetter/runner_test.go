package urlgetter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ooni/probe-engine/experiment/urlgetter"
)

func TestRunnerWithInvalidURLScheme(t *testing.T) {
	r := urlgetter.Runner{Target: "antani://www.google.com"}
	err := r.Run(context.Background())
	if err == nil || err.Error() != "unknown targetURL scheme" {
		t.Fatal("not the error we expected")
	}
}

func TestRunnerHTTPWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := urlgetter.Runner{Target: "https://www.google.com"}
	err := r.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
}

func TestRunnerDNSLookupWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := urlgetter.Runner{Target: "dnslookup://www.google.com"}
	err := r.Run(ctx)
	if err == nil || err.Error() != "interrupted" {
		t.Fatal("not the error we expected")
	}
}

func TestRunnerTLSHandshakeWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := urlgetter.Runner{Target: "tlshandshake://www.google.com:443"}
	err := r.Run(ctx)
	if err == nil || err.Error() != "interrupted" {
		t.Fatal("not the error we expected")
	}
}

func TestRunnerTCPConnectWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := urlgetter.Runner{Target: "tcpconnect://www.google.com:443"}
	err := r.Run(ctx)
	if err == nil || err.Error() != "interrupted" {
		t.Fatal("not the error we expected")
	}
}

func TestRunnerWithInvalidURL(t *testing.T) {
	r := urlgetter.Runner{Target: "\t"}
	err := r.Run(context.Background())
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
}

func TestRunnerWithEmptyHostname(t *testing.T) {
	r := urlgetter.Runner{Target: "http:///foo.txt"}
	err := r.Run(context.Background())
	if err == nil || !strings.HasSuffix(err.Error(), "no Host in request URL") {
		t.Fatal("not the error we expected")
	}
}

func TestIntegrationRunnerTLSHandshakeSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	r := urlgetter.Runner{Target: "tlshandshake://www.google.com:443"}
	err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationRunnerTCPConnectSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	r := urlgetter.Runner{Target: "tcpconnect://www.google.com:443"}
	err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationRunnerDNSLookupSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	r := urlgetter.Runner{Target: "dnslookup://www.google.com"}
	err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationRunnerHTTPSSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	r := urlgetter.Runner{Target: "https://www.google.com"}
	err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunnerHTTPSetHostHeader(t *testing.T) {
	var host string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host = r.Host
		w.WriteHeader(200)
	}))
	defer server.Close()
	r := urlgetter.Runner{
		Config: urlgetter.Config{
			HTTPHost: "x.org",
		},
		Target: server.URL,
	}
	err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if host != "x.org" {
		t.Fatal("not the host we expected")
	}
}

func TestRunnerHTTPNoRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", "http:///") // cause failure if we redirect
		w.WriteHeader(302)
	}))
	defer server.Close()
	r := urlgetter.Runner{
		Config: urlgetter.Config{
			NoFollowRedirects: true,
		},
		Target: server.URL,
	}
	err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
