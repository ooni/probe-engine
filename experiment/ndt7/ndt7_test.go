package ndt7

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestUnitNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "ndt" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.6.0" {
		t.Fatal("unexpected version")
	}
}

func TestUnitDiscoverCancelledContext(t *testing.T) {
	m := new(measurer)
	sess := &mockable.ExperimentSession{
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

func TestUnitDoDownloadWithCancelledContext(t *testing.T) {
	m := new(measurer)
	sess := &mockable.ExperimentSession{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := m.doDownload(
		ctx, sess, handler.NewPrinterCallbacks(log.Log), new(TestKeys),
		"ws://host.name")
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
}

func TestUnitDoUploadWithCancelledContext(t *testing.T) {
	m := new(measurer)
	sess := &mockable.ExperimentSession{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := m.doUpload(
		ctx, sess, handler.NewPrinterCallbacks(log.Log), new(TestKeys),
		"ws://host.name")
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunWithCancelledContext(t *testing.T) {
	m := new(measurer)
	sess := &mockable.ExperimentSession{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := m.Run(ctx, sess, new(model.Measurement), handler.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunWithMaybeStartTunnelFailure(t *testing.T) {
	m := new(measurer)
	expected := errors.New("mocked error")
	sess := &mockable.ExperimentSession{
		MockableHTTPClient:          http.DefaultClient,
		MockableMaybeStartTunnelErr: expected,
		MockableLogger:              log.Log,
		MockableUserAgent:           "miniooni/0.1.0-dev",
	}
	measurement := new(model.Measurement)
	err := m.Run(context.TODO(), sess, measurement, handler.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunWithProxyURL(t *testing.T) {
	m := new(measurer)
	sess := &mockable.ExperimentSession{
		MockableHTTPClient: http.DefaultClient,
		MockableLogger:     log.Log,
		MockableProxyURL:   &url.URL{Host: "1.1.1.1:22"},
		MockableUserAgent:  "miniooni/0.1.0-dev",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	measurement := new(model.Measurement)
	err := m.Run(ctx, sess, measurement, handler.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if measurement.TestKeys.(*TestKeys).SOCKSProxy != "1.1.1.1:22" {
		t.Fatal("not the SOCKSProxy we expected")
	}
}

func TestIntegration(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationFailDownload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	measurer := NewExperimentMeasurer(Config{}).(*measurer)
	measurer.preDownloadHook = func() {
		cancel()
	}
	err := measurer.Run(
		ctx,
		&mockable.ExperimentSession{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal(err)
	}
}

func TestIntegrationFailUpload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	measurer := NewExperimentMeasurer(Config{}).(*measurer)
	measurer.preUploadHook = func() {
		cancel()
	}
	err := measurer.Run(
		ctx,
		&mockable.ExperimentSession{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal(err)
	}
}

func TestIntegrationDownloadJSONUnmarshalFail(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{}).(*measurer)
	var seenError bool
	expected := errors.New("expected error")
	measurer.jsonUnmarshal = func(data []byte, v interface{}) error {
		seenError = true
		return expected
	}
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !seenError {
		t.Fatal("did not see expected error")
	}
}
