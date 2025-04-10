package netxlite

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ooni/probe-engine/pkg/mocks"
	"github.com/ooni/probe-engine/pkg/model"
)

func TestNewHTTPTransportWithOptions(t *testing.T) {

	t.Run("make sure that we get the correct types and settings", func(t *testing.T) {
		expectDialer := &mocks.Dialer{}
		expectTLSDialer := &mocks.TLSDialer{}
		expectLogger := model.DiscardLogger
		txp := NewHTTPTransportWithOptions(expectLogger, expectDialer, expectTLSDialer)

		// undo the results of the netxlite.WrapTransport function
		txpLogger := txp.(*httpTransportLogger)
		if txpLogger.Logger != expectLogger {
			t.Fatal("invalid logger")
		}
		txpErrWrapper := txpLogger.HTTPTransport.(*httpTransportErrWrapper)

		// make sure we correctly configured dialer and TLS dialer
		txpCloser := txpErrWrapper.HTTPTransport.(*httpTransportConnectionsCloser)
		timeoutDialer := txpCloser.Dialer.(*httpDialerWithReadTimeout)
		childDialer := timeoutDialer.Dialer
		if childDialer != expectDialer {
			t.Fatal("invalid dialer")
		}
		timeoutTLSDialer := txpCloser.TLSDialer.(*httpTLSDialerWithReadTimeout)
		childTLSDialer := timeoutTLSDialer.TLSDialer
		if childTLSDialer != expectTLSDialer {
			t.Fatal("invalid TLS dialer")
		}

		// make sure there's the stdlib adapter
		stdlibAdapter := txpCloser.HTTPTransport.(*httpTransportStdlib)
		underlying := stdlibAdapter.StdlibTransport

		// now let's check that everything is configured as intended
		expectedTxp := http.DefaultTransport.(*http.Transport).Clone()
		diff := cmp.Diff(
			expectedTxp,
			underlying,
			cmpopts.IgnoreUnexported(http.Transport{}),
			cmpopts.IgnoreUnexported(tls.Config{}),
			cmpopts.IgnoreFields(
				http.Transport{},
				"DialContext",
				"DialTLSContext",
				"DisableCompression",
				"Proxy",
				"ForceAttemptHTTP2",
			),
		)
		if diff != "" {
			t.Fatal(diff)
		}

		// finish checking by explicitly inspecting the fields we modify
		if underlying.DialContext == nil {
			t.Fatal("expected non-nil .DialContext")
		}
		if underlying.DialTLSContext == nil {
			t.Fatal("expected non-nil .DialTLSContext")
		}
		if underlying.Proxy != nil {
			t.Fatal("expected nil .Proxy")
		}
		if !underlying.ForceAttemptHTTP2 {
			t.Fatal("expected true .ForceAttemptHTTP2")
		}
		if !underlying.DisableCompression {
			t.Fatal("expected true .DisableCompression")
		}
	})

	unwrap := func(txp model.HTTPTransport) *http.Transport {
		txpLogger := txp.(*httpTransportLogger)
		txpErrWrapper := txpLogger.HTTPTransport.(*httpTransportErrWrapper)
		txpCloser := txpErrWrapper.HTTPTransport.(*httpTransportConnectionsCloser)
		stdlibAdapter := txpCloser.HTTPTransport.(*httpTransportStdlib)
		return stdlibAdapter.StdlibTransport
	}

	t.Run("make sure HTTPTransportOptionProxyURL is WAI", func(t *testing.T) {
		runWithURL := func(expectedURL *url.URL) {
			expectDialer := &mocks.Dialer{}
			expectTLSDialer := &mocks.TLSDialer{}
			expectLogger := model.DiscardLogger
			txp := NewHTTPTransportWithOptions(
				expectLogger,
				expectDialer,
				expectTLSDialer,
				HTTPTransportOptionProxyURL(expectedURL),
			)
			underlying := unwrap(txp)
			if underlying.Proxy == nil {
				t.Fatal("expected non-nil .Proxy")
			}
			got, err := underlying.Proxy(&http.Request{})
			if err != nil {
				t.Fatal(err)
			}
			if got != expectedURL {
				t.Fatal("not the expected URL")
			}
		}

		runWithURL(&url.URL{})

		runWithURL(nil)
	})

	t.Run("make sure HTTPTransportOptionMaxConnsPerHost is WAI", func(t *testing.T) {
		runWithValue := func(expectedValue int) {
			expectDialer := &mocks.Dialer{}
			expectTLSDialer := &mocks.TLSDialer{}
			expectLogger := model.DiscardLogger
			txp := NewHTTPTransportWithOptions(
				expectLogger,
				expectDialer,
				expectTLSDialer,
				HTTPTransportOptionMaxConnsPerHost(expectedValue),
			)
			underlying := unwrap(txp)
			got := underlying.MaxConnsPerHost
			if got != expectedValue {
				t.Fatal("not the expected value")
			}
		}

		runWithValue(100)

		runWithValue(10)
	})

	t.Run("make sure HTTPTransportDisableCompression is WAI", func(t *testing.T) {
		runWithValue := func(expectedValue bool) {
			expectDialer := &mocks.Dialer{}
			expectTLSDialer := &mocks.TLSDialer{}
			expectLogger := model.DiscardLogger
			txp := NewHTTPTransportWithOptions(
				expectLogger,
				expectDialer,
				expectTLSDialer,
				HTTPTransportOptionDisableCompression(expectedValue),
			)
			underlying := unwrap(txp)
			got := underlying.DisableCompression
			if got != expectedValue {
				t.Fatal("not the expected value")
			}
		}

		runWithValue(true)

		runWithValue(false)
	})
}
