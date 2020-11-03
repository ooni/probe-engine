package example_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestSuccess(t *testing.T) {
	m := example.NewExperimentMeasurer(example.Config{
		SleepTime: int64(2 * time.Millisecond),
	}, "example")
	if m.ExperimentName() != "example" {
		t.Fatal("invalid ExperimentName")
	}
	if m.ExperimentVersion() != "0.0.1" {
		t.Fatal("invalid ExperimentVersion")
	}
	ctx := context.Background()
	sess := &mockable.Session{MockableLogger: log.Log}
	callbacks := model.NewPrinterCallbacks(sess.Logger())
	err := m.Run(ctx, sess, new(model.Measurement), callbacks)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFailure(t *testing.T) {
	m := example.NewExperimentMeasurer(example.Config{
		SleepTime:   int64(2 * time.Millisecond),
		ReturnError: true,
	}, "example")
	ctx := context.Background()
	sess := &mockable.Session{MockableLogger: log.Log}
	callbacks := model.NewPrinterCallbacks(sess.Logger())
	err := m.Run(ctx, sess, new(model.Measurement), callbacks)
	if !errors.Is(err, example.ErrFailure) {
		t.Fatal("expected an error here")
	}
}
