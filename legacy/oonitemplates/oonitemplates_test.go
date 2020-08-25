package oonitemplates

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	goptlib "git.torproject.org/pluggable-transports/goptlib.git"
	"github.com/ooni/probe-engine/netx/modelx"
	"gitlab.com/yawning/obfs4.git/transports"
	obfs4base "gitlab.com/yawning/obfs4.git/transports/base"
)

func TestUnitChannelHandlerWriteLateOnChannel(t *testing.T) {
	handler := newChannelHandler(make(chan modelx.Measurement))
	var waitgroup sync.WaitGroup
	waitgroup.Add(1)
	go func() {
		time.Sleep(1 * time.Second)
		handler.OnMeasurement(modelx.Measurement{})
		waitgroup.Done()
	}()
	waitgroup.Wait()
	if handler.lateWrites.Load() != 1 {
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
	if results.Error.Error() != modelx.FailureGenericTimeoutError {
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
	if results.Error.Error() != modelx.FailureGenericTimeoutError {
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

func obfs4config() OBFS4ConnectConfig {
	// TODO(bassosimone): this is a public working bridge we have found
	// with @hellais. We should ask @phw whether there is some obfs4 bridge
	// dedicated to integration testing that we should use instead.
	return OBFS4ConnectConfig{
		Address:      "109.105.109.165:10527",
		StateBaseDir: "../../testdata/",
		Params: map[string][]string{
			"cert": []string{
				"Bvg/itxeL4TWKLP6N1MaQzSOC6tcRIBv6q57DYAZc3b2AzuM+/TfB7mqTFEfXILCjEwzVA",
			},
			"iat-mode": []string{"1"},
		},
	}
}

func TestIntegrationOBFS4ConnectGood(t *testing.T) {
	ctx := context.Background()
	results := OBFS4Connect(ctx, obfs4config())
	if results.Error != nil {
		t.Fatal(results.Error)
	}
}

func TestIntegrationOBFS4ConnectGoodWithDoT(t *testing.T) {
	ctx := context.Background()
	config := obfs4config()
	config.DNSServerNetwork = "dot"
	config.DNSServerAddress = "9.9.9.9:853"
	results := OBFS4Connect(ctx, config)
	if results.Error != nil {
		t.Fatal(results.Error)
	}
}

func TestIntegrationOBFS4ConnectUnknownDNS(t *testing.T) {
	ctx := context.Background()
	config := obfs4config()
	config.DNSServerNetwork = "antani"
	results := OBFS4Connect(ctx, config)
	if !strings.HasSuffix(results.Error.Error(), "unsupported network value") {
		t.Fatal("not the error that we expected")
	}
}

func TestIntegrationOBFS4IoutilTempDirError(t *testing.T) {
	ctx := context.Background()
	config := obfs4config()
	expected := errors.New("mocked error")
	config.ioutilTempDir = func(dir, prefix string) (string, error) {
		return "", expected
	}
	results := OBFS4Connect(ctx, config)
	if !errors.Is(results.Error, expected) {
		t.Fatal("not the error that we expected")
	}
}

func TestIntegrationOBFS4ClientFactoryError(t *testing.T) {
	ctx := context.Background()
	config := obfs4config()
	config.transportsGet = func(name string) obfs4base.Transport {
		txp := transports.Get(name)
		if name == "obfs4" && txp != nil {
			txp = &faketransport{txp: txp}
		}
		return txp
	}
	results := OBFS4Connect(ctx, config)
	if results.Error.Error() != "mocked ClientFactory error" {
		t.Fatal("not the error we expected")
	}
}

func TestIntegrationOBFS4ParseArgsError(t *testing.T) {
	ctx := context.Background()
	config := obfs4config()
	config.Params = make(map[string][]string) // cause ParseArgs error
	results := OBFS4Connect(ctx, config)
	if results.Error.Error() != "missing argument 'node-id'" {
		t.Fatal("not the error we expected")
	}
}

func TestIntegrationOBFS4DialContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // should cause DialContex to fail
	config := obfs4config()
	results := OBFS4Connect(ctx, config)
	if results.Error.Error() != "interrupted" {
		t.Fatal("not the error we expected")
	}
}

func TestIntegrationOBFS4SetDeadlineError(t *testing.T) {
	ctx := context.Background()
	config := obfs4config()
	config.setDeadline = func(net.Conn, time.Time) error {
		return errors.New("mocked error")
	}
	results := OBFS4Connect(ctx, config)
	if !strings.HasSuffix(results.Error.Error(), "mocked error") {
		t.Fatal("not the error we expected")
	}
}

type faketransport struct {
	txp obfs4base.Transport
}

func (txp *faketransport) Name() string {
	return txp.Name()
}

func (txp *faketransport) ClientFactory(stateDir string) (obfs4base.ClientFactory, error) {
	return nil, errors.New("mocked ClientFactory error")
}

func (txp *faketransport) ServerFactory(stateDir string, args *goptlib.Args) (obfs4base.ServerFactory, error) {
	return txp.ServerFactory(stateDir, args)
}

func TestUnitConnmapper(t *testing.T) {
	var mapper connmapper
	if mapper.scramble(-1) >= 0 {
		t.Fatal("unexpected value for negative input")
	}
	if mapper.scramble(1234) != 2 {
		t.Fatal("unexpected second value")
	}
	if mapper.scramble(12) != 3 {
		t.Fatal("unexpected third value")
	}
	if mapper.scramble(12) != mapper.scramble(12) {
		t.Fatal("not idempotent")
	}
	if mapper.scramble(0) != 0 {
		t.Fatal("unexpected value for zero input")
	}
}
