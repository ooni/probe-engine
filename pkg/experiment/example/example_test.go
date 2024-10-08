package example_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/experiment/example"
	"github.com/ooni/probe-engine/pkg/mocks"
	"github.com/ooni/probe-engine/pkg/model"
)

func TestSuccess(t *testing.T) {
	m := example.NewExperimentMeasurer(example.Config{
		SleepTime: int64(2 * time.Millisecond),
	})
	if m.ExperimentName() != "example" {
		t.Fatal("invalid ExperimentName")
	}
	if m.ExperimentVersion() != "0.1.0" {
		t.Fatal("invalid ExperimentVersion")
	}
	ctx := context.Background()
	sess := &mocks.Session{
		MockLogger: func() model.Logger {
			return log.Log
		},
	}
	callbacks := model.NewPrinterCallbacks(sess.Logger())
	measurement := new(model.Measurement)
	args := &model.ExperimentArgs{
		Callbacks:   callbacks,
		Measurement: measurement,
		Session:     sess,
	}
	err := m.Run(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFailure(t *testing.T) {
	m := example.NewExperimentMeasurer(example.Config{
		SleepTime:   int64(2 * time.Millisecond),
		ReturnError: true,
	})
	ctx := context.Background()
	sess := &mocks.Session{
		MockLogger: func() model.Logger {
			return log.Log
		},
	}
	callbacks := model.NewPrinterCallbacks(sess.Logger())
	args := &model.ExperimentArgs{
		Callbacks:   callbacks,
		Measurement: new(model.Measurement),
		Session:     sess,
	}
	err := m.Run(ctx, args)
	if !errors.Is(err, example.ErrFailure) {
		t.Fatal("expected an error here")
	}
}
