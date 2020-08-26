package psiphon_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/experiment/psiphon"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

// Implementation note: integration test performed by
// the $topdir/experiment_test.go file

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := psiphon.NewExperimentMeasurer(psiphon.Config{})
	if measurer.ExperimentName() != "psiphon" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.4.0" {
		t.Fatal("unexpected version")
	}
}

func TestRunWithCancelledContext(t *testing.T) {
	measurer := psiphon.NewExperimentMeasurer(psiphon.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	measurement := new(model.Measurement)
	err := measurer.Run(ctx, newfakesession(), measurement,
		model.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("expected another error here")
	}
	tk := measurement.TestKeys.(psiphon.TestKeys)
	if tk.MaxRuntime <= 0 {
		t.Fatal("you did not set the max runtime")
	}
}

func TestRunWithCustomInputAndCancelledContext(t *testing.T) {
	expected := "http://x.org"
	measurement := &model.Measurement{
		Input: model.MeasurementTarget(expected),
	}
	measurer := psiphon.NewExperimentMeasurer(psiphon.Config{})
	measurer.(*psiphon.Measurer).BeforeGetHook = func(g urlgetter.Getter) {
		if g.Target != expected {
			t.Fatal("target was not correctly set")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	err := measurer.Run(ctx, newfakesession(), measurement,
		model.NewPrinterCallbacks(log.Log))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("expected another error here")
	}
	tk := measurement.TestKeys.(psiphon.TestKeys)
	if tk.MaxRuntime <= 0 {
		t.Fatal("you did not set the max runtime")
	}
}

func TestRunWillPrintSomethingWithCancelledContext(t *testing.T) {
	measurement := new(model.Measurement)
	measurer := psiphon.NewExperimentMeasurer(psiphon.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	measurer.(*psiphon.Measurer).BeforeGetHook = func(g urlgetter.Getter) {
		time.Sleep(2 * time.Second)
		cancel() // fail after we've given the printer a chance to run
	}
	observer := observerCallbacks{progress: atomicx.NewInt64()}
	err := measurer.Run(ctx, newfakesession(), measurement, observer)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("expected another error here")
	}
	tk := measurement.TestKeys.(psiphon.TestKeys)
	if tk.MaxRuntime <= 0 {
		t.Fatal("you did not set the max runtime")
	}
	if observer.progress.Load() < 2 {
		t.Fatal("not enough progress emitted?!")
	}
}

type observerCallbacks struct {
	progress *atomicx.Int64
}

func (d observerCallbacks) OnDataUsage(dloadKiB, uploadKiB float64) {
}

func (d observerCallbacks) OnProgress(percentage float64, message string) {
	d.progress.Add(1)
}

func newfakesession() model.ExperimentSession {
	return &mockable.ExperimentSession{MockableLogger: log.Log}
}
