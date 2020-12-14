package run_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/dnscheck"
	"github.com/ooni/probe-engine/experiment/run"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestExperimentNameAndVersion(t *testing.T) {
	measurer := run.NewExperimentMeasurer(run.Config{})
	if measurer.ExperimentName() != "run" {
		t.Error("unexpected experiment name")
	}
	if measurer.ExperimentVersion() != "0.2.0" {
		t.Error("unexpected experiment version")
	}
}

func TestRunDNSCheckWithCancelledContext(t *testing.T) {
	measurer := run.NewExperimentMeasurer(run.Config{})
	input := `{"name": "dnscheck", "input": "https://dns.google/dns-query"}`
	measurement := new(model.Measurement)
	measurement.Input = model.MeasurementTarget(input)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	sess := &mockable.Session{MockableLogger: log.Log}
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	// TODO(bassosimone): here we could improve the tests by checking
	// whether the result makes sense for a cancelled context.
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := measurement.TestKeys.(*dnscheck.TestKeys); !ok {
		t.Fatal("invalid type for test keys")
	}
	sk, err := measurer.GetSummaryKeys(measurement)
	if err != nil {
		t.Fatal(err)
	}
	rsk, ok := sk.(dnscheck.SummaryKeys)
	if !ok {
		t.Fatal("cannot convert summary keys to specific type")
	}
	if rsk.IsAnomaly != false {
		t.Fatal("unexpected IsAnomaly value")
	}
}

func TestRunWithInvalidJSON(t *testing.T) {
	measurer := run.NewExperimentMeasurer(run.Config{})
	input := `{"name": }`
	measurement := new(model.Measurement)
	measurement.Input = model.MeasurementTarget(input)
	ctx := context.Background()
	sess := &mockable.Session{MockableLogger: log.Log}
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err == nil || err.Error() != "invalid character '}' looking for beginning of value" {
		t.Fatalf("not the error we expected: %+v", err)
	}
}

func TestRunWithUnknownExperiment(t *testing.T) {
	measurer := run.NewExperimentMeasurer(run.Config{})
	input := `{"name": "antani"}`
	measurement := new(model.Measurement)
	measurement.Input = model.MeasurementTarget(input)
	ctx := context.Background()
	sess := &mockable.Session{MockableLogger: log.Log}
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err == nil || err.Error() != "no such experiment: antani" {
		t.Fatalf("not the error we expected: %+v", err)
	}
}
