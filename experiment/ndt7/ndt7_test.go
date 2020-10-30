package ndt7

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "ndt" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.6.0" {
		t.Fatal("unexpected version")
	}
}

func TestDiscoverCancelledContext(t *testing.T) {
	m := new(Measurer)
	sess := &mockable.Session{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	locateResult, err := m.discover(ctx, sess)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if locateResult.Hostname != "" {
		t.Fatal("not the Hostname we expected")
	}
}

type verifyRequestTransport struct {
	ExpectedError error
}

func (txp *verifyRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.RawQuery != "ip=1.2.3.4" {
		return nil, errors.New("invalid req.URL.RawQuery")
	}
	return nil, txp.ExpectedError
}

func TestDoDownloadWithCancelledContext(t *testing.T) {
	m := new(Measurer)
	sess := &mockable.Session{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := m.doDownload(
		ctx, sess, model.NewPrinterCallbacks(log.Log), new(TestKeys),
		"ws://host.name")
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
}

func TestDoUploadWithCancelledContext(t *testing.T) {
	m := new(Measurer)
	sess := &mockable.Session{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := m.doUpload(
		ctx, sess, model.NewPrinterCallbacks(log.Log), new(TestKeys),
		"ws://host.name")
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
}

func TestRunWithCancelledContext(t *testing.T) {
	m := new(Measurer)
	sess := &mockable.Session{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := m.Run(ctx, sess, new(model.Measurement), model.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
}

func TestRunWithMaybeStartTunnelFailure(t *testing.T) {
	m := new(Measurer)
	expected := errors.New("mocked error")
	sess := &mockable.Session{
		MockableHTTPClient:          http.DefaultClient,
		MockableMaybeStartTunnelErr: expected,
		MockableLogger:              log.Log,
		MockableUserAgent:           "miniooni/0.1.0-dev",
	}
	measurement := new(model.Measurement)
	err := m.Run(context.TODO(), sess, measurement, model.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestRunWithProxyURL(t *testing.T) {
	m := new(Measurer)
	sess := &mockable.Session{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableProxyURL:   &url.URL{Host: "1.1.1.1:22"},
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	measurement := new(model.Measurement)
	err := m.Run(ctx, sess, measurement, model.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if measurement.TestKeys.(*TestKeys).SOCKSProxy != "1.1.1.1:22" {
		t.Fatal("not the SOCKSProxy we expected")
	}
}

func TestGood(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	measurer := NewExperimentMeasurer(Config{})
	err := measurer.Run(
		context.Background(),
		&mockable.Session{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		model.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFailDownload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	measurer := NewExperimentMeasurer(Config{}).(*Measurer)
	measurer.preDownloadHook = func() {
		cancel()
	}
	err := measurer.Run(
		ctx,
		&mockable.Session{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		model.NewPrinterCallbacks(log.Log),
	)
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal(err)
	}
}

func TestFailUpload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	measurer := NewExperimentMeasurer(Config{noDownload: true}).(*Measurer)
	measurer.preUploadHook = func() {
		cancel()
	}
	err := measurer.Run(
		ctx,
		&mockable.Session{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		model.NewPrinterCallbacks(log.Log),
	)
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal(err)
	}
}

func TestDownloadJSONUnmarshalFail(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{noUpload: true}).(*Measurer)
	var seenError bool
	expected := errors.New("expected error")
	measurer.jsonUnmarshal = func(data []byte, v interface{}) error {
		seenError = true
		return expected
	}
	err := measurer.Run(
		context.Background(),
		&mockable.Session{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		model.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !seenError {
		t.Fatal("did not see expected error")
	}
}
