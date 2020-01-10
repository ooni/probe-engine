package oonitemplates

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ooni/netx/handlers"
	"github.com/ooni/netx/modelx"
	"github.com/ooni/netx/x/scoreboard"
)

func TestUnitChannelHandlerWriteLateOnChannel(t *testing.T) {
	handler := &channelHandler{
		ch: make(chan modelx.Measurement),
	}
	var waitgroup sync.WaitGroup
	waitgroup.Add(1)
	go func() {
		time.Sleep(1 * time.Second)
		handler.OnMeasurement(modelx.Measurement{})
		waitgroup.Done()
	}()
	waitgroup.Wait()
	if handler.lateWrites != 1 {
		t.Fatal("unexpected lateWrites value")
	}
}

func TestIntegrationDNSLookupGood(t *testing.T) {
	ctx := context.Background()
	results := DNSLookup(ctx, DNSLookupConfig{
		Hostname: "ooni.io",
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
	if len(results.Addresses) < 1 {
		t.Fatal("no addresses returned?!")
	}
	if results.TestKeys.Scoreboard == nil {
		t.Fatal("no scoreboard?!")
	}
}

func TestIntegrationDNSLookupCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Microsecond,
	)
	defer cancel()
	results := DNSLookup(ctx, DNSLookupConfig{
		Hostname: "ooni.io",
	})
	if results.Error == nil {
		t.Fatal("expected an error here")
	}
	if results.Error.Error() != "generic_timeout_error" {
		t.Fatal("not the error we expected")
	}
	if len(results.Addresses) > 0 {
		t.Fatal("addresses returned?!")
	}
}

func TestIntegrationDNSLookupUnknownDNS(t *testing.T) {
	ctx := context.Background()
	results := DNSLookup(ctx, DNSLookupConfig{
		Hostname:      "ooni.io",
		ServerNetwork: "antani",
	})
	if !strings.HasSuffix(results.Error.Error(), "unsupported network value") {
		t.Fatal("expected a different error here")
	}
}

func TestIntegrationHTTPDoGood(t *testing.T) {
	ctx := context.Background()
	results := HTTPDo(ctx, HTTPDoConfig{
		Accept:         "*/*",
		AcceptLanguage: "en",
		URL:            "http://ooni.io",
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
	if results.StatusCode != 200 {
		t.Fatal("request failed?!")
	}
	if len(results.Headers) < 1 {
		t.Fatal("no headers?!")
	}
	if len(results.BodySnap) < 1 {
		t.Fatal("no body?!")
	}
	if results.TestKeys.Scoreboard == nil {
		t.Fatal("no scoreboard?!")
	}
}

func TestIntegrationHTTPDoUnknownDNS(t *testing.T) {
	ctx := context.Background()
	results := HTTPDo(ctx, HTTPDoConfig{
		URL:              "http://ooni.io",
		DNSServerNetwork: "antani",
	})
	if !strings.HasSuffix(results.Error.Error(), "unsupported network value") {
		t.Fatal("not the error that we expected")
	}
}

func TestIntegrationHTTPDoForceSkipVerify(t *testing.T) {
	ctx := context.Background()
	results := HTTPDo(ctx, HTTPDoConfig{
		URL:                "https://self-signed.badssl.com/",
		InsecureSkipVerify: true,
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
}

func TestIntegrationHTTPDoRoundTripError(t *testing.T) {
	ctx := context.Background()
	results := HTTPDo(ctx, HTTPDoConfig{
		URL: "http://ooni.io:443", // 443 with http
	})
	if results.Error == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationHTTPDoBadURL(t *testing.T) {
	ctx := context.Background()
	results := HTTPDo(ctx, HTTPDoConfig{
		URL: "\t",
	})
	if !strings.HasSuffix(results.Error.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
}

func TestIntegrationTLSConnectGood(t *testing.T) {
	ctx := context.Background()
	results := TLSConnect(ctx, TLSConnectConfig{
		Address: "ooni.io:443",
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
	if results.TestKeys.Scoreboard == nil {
		t.Fatal("no scoreboard?!")
	}
}

func TestIntegrationTLSConnectGoodWithDoT(t *testing.T) {
	ctx := context.Background()
	results := TLSConnect(ctx, TLSConnectConfig{
		Address:          "ooni.io:443",
		DNSServerNetwork: "dot",
		DNSServerAddress: "9.9.9.9:853",
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
	if results.TestKeys.Scoreboard == nil {
		t.Fatal("no scoreboard?!")
	}
}

func TestIntegrationTLSConnectCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Microsecond,
	)
	defer cancel()
	results := TLSConnect(ctx, TLSConnectConfig{
		Address: "ooni.io:443",
	})
	if results.Error == nil {
		t.Fatal("expected an error here")
	}
	if results.Error.Error() != "generic_timeout_error" {
		t.Fatal("not the error we expected")
	}
}

func TestIntegrationTLSConnectUnknownDNS(t *testing.T) {
	ctx := context.Background()
	results := TLSConnect(ctx, TLSConnectConfig{
		Address:          "ooni.io:443",
		DNSServerNetwork: "antani",
	})
	if !strings.HasSuffix(results.Error.Error(), "unsupported network value") {
		t.Fatal("not the error that we expected")
	}
}

func TestIntegrationTLSConnectForceSkipVerify(t *testing.T) {
	ctx := context.Background()
	results := TLSConnect(ctx, TLSConnectConfig{
		Address:            "self-signed.badssl.com:443",
		InsecureSkipVerify: true,
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
}

func TestMaybeRunTLSChecks(t *testing.T) {
	out := maybeRunTLSChecks(
		context.Background(), handlers.NoHandler,
		&modelx.XResults{
			Scoreboard: scoreboard.Board{
				TLSHandshakeReset: []scoreboard.TLSHandshakeReset{
					scoreboard.TLSHandshakeReset{
						Domain: "ooni.io",
						RecommendedFollowups: []string{
							"sni_blocking",
						},
					},
				},
			},
		},
	)
	if out == nil {
		t.Fatal("unexpected nil return value")
	}
	if out.Connects == nil {
		t.Fatal("no connects?!")
	}
	if out.HTTPRequests != nil {
		t.Fatal("http requests?!")
	}
	if out.Resolves == nil {
		t.Fatal("no queries?!")
	}
	if out.TLSHandshakes == nil {
		t.Fatal("no TLS handshakes?!")
	}
}

func TestIntegrationBodySnapSizes(t *testing.T) {
	const (
		maxEventsBodySnapSize   = 1 << 7
		maxResponseBodySnapSize = 1 << 8
	)
	ctx := context.Background()
	results := HTTPDo(ctx, HTTPDoConfig{
		URL:                     "https://ooni.org",
		MaxEventsBodySnapSize:   maxEventsBodySnapSize,
		MaxResponseBodySnapSize: maxResponseBodySnapSize,
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
	if results.StatusCode != 200 {
		t.Fatal("request failed?!")
	}
	if len(results.Headers) < 1 {
		t.Fatal("no headers?!")
	}
	if len(results.BodySnap) != maxResponseBodySnapSize {
		t.Fatal("invalid response body snap size")
	}
	if results.TestKeys.Scoreboard == nil {
		t.Fatal("no scoreboard?!")
	}
	if results.TestKeys.HTTPRequests == nil {
		t.Fatal("no HTTPRequests?!")
	}
	for _, req := range results.TestKeys.HTTPRequests {
		if len(req.ResponseBodySnap) != maxEventsBodySnapSize {
			t.Fatal("invalid length of ResponseBodySnap")
		}
		if req.MaxBodySnapSize != maxEventsBodySnapSize {
			t.Fatal("unexpected value of MaxBodySnapSize")
		}
	}
}

func TestIntegrationTCPConnectGood(t *testing.T) {
	ctx := context.Background()
	results := TCPConnect(ctx, TCPConnectConfig{
		Address: "ooni.io:443",
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
	if results.TestKeys.Scoreboard == nil {
		t.Fatal("no scoreboard?!")
	}
}

func TestIntegrationTCPConnectGoodWithDoT(t *testing.T) {
	ctx := context.Background()
	results := TCPConnect(ctx, TCPConnectConfig{
		Address:          "ooni.io:443",
		DNSServerNetwork: "dot",
		DNSServerAddress: "9.9.9.9:853",
	})
	if results.Error != nil {
		t.Fatal(results.Error)
	}
	if results.TestKeys.Scoreboard == nil {
		t.Fatal("no scoreboard?!")
	}
}

func TestIntegrationTCPConnectUnknownDNS(t *testing.T) {
	ctx := context.Background()
	results := TCPConnect(ctx, TCPConnectConfig{
		Address:          "ooni.io:443",
		DNSServerNetwork: "antani",
	})
	if !strings.HasSuffix(results.Error.Error(), "unsupported network value") {
		t.Fatal("not the error that we expected")
	}
}
