package sniblocking

import (
	"context"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/netxlogger"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestUnitTestKeysClassify(t *testing.T) {
	asStringPtr := func(s string) *string {
		return &s
	}
	t.Run("with tk.Target.Failure == nil", func(t *testing.T) {
		tk := new(TestKeys)
		if tk.classify() != classAccessibleValidHostname {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == connection_refused", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureConnectionRefused)
		if tk.classify() != classAnomalyTestHelperBlocked {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == dns_nxdomain_error", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureDNSNXDOMAINError)
		if tk.classify() != classAnomalyTestHelperBlocked {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == connection_reset", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureConnectionReset)
		if tk.classify() != classBlockedTCPIPError {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == eof_error", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureEOFError)
		if tk.classify() != classBlockedTCPIPError {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == ssl_invalid_hostname", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureSSLInvalidHostname)
		if tk.classify() != classAccessibleInvalidHostname {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == ssl_unknown_authority", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureSSLUnknownAuthority)
		if tk.classify() != classAnomalySSLError {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == ssl_invalid_certificate", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureSSLInvalidCertificate)
		if tk.classify() != classAnomalySSLError {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == generic_timeout_error #1", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureGenericTimeoutError)
		if tk.classify() != classAnomalyTimeout {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == generic_timeout_error #2", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr(modelx.FailureGenericTimeoutError)
		tk.Control.Failure = asStringPtr(modelx.FailureGenericTimeoutError)
		if tk.classify() != classAnomalyTestHelperBlocked {
			t.Fatal("unexpected result")
		}
	})
	t.Run("with tk.Target.Failure == unknown_failure", func(t *testing.T) {
		tk := new(TestKeys)
		tk.Target.Failure = asStringPtr("unknown_failure")
		if tk.classify() != classAnomalyUnexpectedFailure {
			t.Fatal("unexpected result")
		}
	})
}

func TestUnitNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "sni_blocking" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.0.4" {
		t.Fatal("unexpected version")
	}
}

func TestUnitMeasurerMeasureNoControlSNI(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	err := measurer.Run(
		context.Background(),
		newsession(),
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err.Error() != "Experiment requires ControlSNI" {
		t.Fatal("not the error we expected")
	}
}

func TestUnitMeasurerMeasureNoMeasurementInput(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{
		ControlSNI: "example.com",
	})
	err := measurer.Run(
		context.Background(),
		newsession(),
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err.Error() != "Experiment requires measurement.Input" {
		t.Fatal("not the error we expected")
	}
}

func TestUnitMeasurerMeasureWithInvalidInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel the context
	measurer := NewExperimentMeasurer(Config{
		ControlSNI: "example.com",
	})
	measurement := &model.Measurement{
		Input: "\t",
	}
	err := measurer.Run(
		ctx,
		newsession(),
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestUnitMeasurerMeasureWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel the context
	measurer := NewExperimentMeasurer(Config{
		ControlSNI: "example.com",
	})
	measurement := &model.Measurement{
		Input: "kernel.org",
	}
	err := measurer.Run(
		ctx,
		newsession(),
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnitMeasureoneCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel the context
	result := new(measurer).measureone(
		ctx,
		netxlogger.NewHandler(log.Log),
		time.Now(),
		"kernel.org",
		"example.com:443",
	)
	if *result.Failure != modelx.FailureGenericTimeoutError {
		t.Fatal("unexpected failure")
	}
	if result.SNI != "kernel.org" {
		t.Fatal("unexpected SNI")
	}
	if result.BytesReceived != 0 {
		t.Fatal("expected to receive bytes")
	}
	if result.BytesSent != 0 {
		t.Fatal("expected to send bytes")
	}
}

func TestUnitMeasureoneSuccess(t *testing.T) {
	result := new(measurer).measureone(
		context.Background(),
		netxlogger.NewHandler(log.Log),
		time.Now(),
		"kernel.org",
		"example.com:443",
	)
	if *result.Failure != modelx.FailureSSLInvalidHostname {
		t.Fatal("unexpected failure")
	}
	if result.SNI != "kernel.org" {
		t.Fatal("unexpected SNI")
	}
	if result.BytesReceived <= 0 {
		t.Fatal("expected to receive bytes")
	}
	if result.BytesSent <= 0 {
		t.Fatal("expected to send bytes")
	}
}

func TestUnitMeasureonewithcacheWorks(t *testing.T) {
	measurer := &measurer{cache: make(map[string]Subresult)}
	output := make(chan Subresult, 2)
	for i := 0; i < 2; i++ {
		measurer.measureonewithcache(
			context.Background(),
			output,
			netxlogger.NewHandler(log.Log),
			time.Now(),
			"kernel.org",
			"example.com:443",
		)
	}
	for _, expected := range []bool{false, true} {
		result := <-output
		if result.Cached != expected {
			t.Fatal("unexpected cached")
		}
		if *result.Failure != modelx.FailureSSLInvalidHostname {
			t.Fatal("unexpected failure")
		}
		if result.SNI != "kernel.org" {
			t.Fatal("unexpected SNI")
		}
		if result.BytesReceived <= 0 && !result.Cached {
			t.Fatal("expected to receive bytes")
		}
		if result.BytesSent <= 0 && !result.Cached {
			t.Fatal("expected to send bytes")
		}
		if result.BytesReceived != 0 && result.Cached {
			t.Fatal("expected to not receive bytes")
		}
		if result.BytesSent != 0 && result.Cached {
			t.Fatal("expected to not send bytes")
		}
	}
}

func TestUnitProcessallPanicsIfInvalidSNI(t *testing.T) {
	defer func() {
		panicdata := recover()
		if panicdata == nil {
			t.Fatal("expected to see panic here")
		}
		if panicdata.(string) != "unexpected smk.SNI" {
			t.Fatal("not the panic we expected")
		}
	}()
	outputs := make(chan Subresult, 1)
	measurement := &model.Measurement{
		Input: "kernel.org",
	}
	go func() {
		outputs <- Subresult{
			SNI: "antani.io",
		}
	}()
	processall(
		outputs,
		measurement,
		handler.NewPrinterCallbacks(log.Log),
		[]string{"kernel.org", "example.com"},
		newsession(),
		"example.com",
	)
}

func TestUnitMaybeURLToSNI(t *testing.T) {
	t.Run("for invalid URL", func(t *testing.T) {
		parsed, err := maybeURLToSNI("\t")
		if err == nil {
			t.Fatal("expected an error here")
		}
		if parsed != "" {
			t.Fatal("expected empty parsed here")
		}
	})
	t.Run("for domain name", func(t *testing.T) {
		parsed, err := maybeURLToSNI("kernel.org")
		if err != nil {
			t.Fatal(err)
		}
		if parsed != "kernel.org" {
			t.Fatal("expected different domain here")
		}
	})
	t.Run("for valid URL", func(t *testing.T) {
		parsed, err := maybeURLToSNI("https://kernel.org/robots.txt")
		if err != nil {
			t.Fatal(err)
		}
		if parsed != "kernel.org" {
			t.Fatal("expected different domain here")
		}
	})
}

func newsession() model.ExperimentSession {
	return &mockable.ExperimentSession{MockableLogger: log.Log}
}
