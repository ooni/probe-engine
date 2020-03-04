package sniblocking

import (
	"context"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/netxlogger"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestUnitNewExperiment(t *testing.T) {
	experiment := NewExperiment(newsession(), Config{})
	if experiment == nil {
		t.Fatal("nil experiment returned")
	}
}

func TestUnitMeasurerMeasureNoControlSNI(t *testing.T) {
	measurer := newMeasurer(Config{})
	err := measurer.measure(
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
	measurer := newMeasurer(Config{
		ControlSNI: "ps.ooni.io",
	})
	err := measurer.measure(
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
	measurer := newMeasurer(Config{
		ControlSNI: "ps.ooni.io",
	})
	measurement := &model.Measurement{
		Input: "\t",
	}
	err := measurer.measure(
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
	measurer := newMeasurer(Config{
		ControlSNI: "ps.ooni.io",
	})
	measurement := &model.Measurement{
		Input: "kernel.org",
	}
	err := measurer.measure(
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
	outputs := make(chan Subresult, 1)
	measureone(
		ctx,
		outputs,
		netxlogger.NewHandler(log.Log),
		time.Now(),
		"kernel.org",
		"ps.ooni.io:443",
	)
	for result := range outputs {
		if *result.Failure != "generic_timeout_error" {
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
		break
	}
}

func TestUnitMeasureoneSuccess(t *testing.T) {
	outputs := make(chan Subresult, 1)
	measureone(
		context.Background(),
		outputs,
		netxlogger.NewHandler(log.Log),
		time.Now(),
		"kernel.org",
		"ps.ooni.io:443",
	)
	for result := range outputs {
		if *result.Failure != "ssl_invalid_hostname" {
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
		break
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
		[]string{"kernel.org", "ps.ooni.io"},
		newsession(),
		"ps.ooni.io",
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

func newsession() *session.Session {
	return session.New(
		log.Log, softwareName, softwareVersion,
		"../../testdata", nil, "../../testdata",
		kvstore.NewMemoryKeyValueStore(),
	)
}
