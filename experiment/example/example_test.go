package example_test

import (
	"context"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/internal/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestIntegrationSuccess(t *testing.T) {
	m := example.NewExperimentMeasurer(example.Config{
		SleepTime: int64(2 * time.Second),
	}, "example")
	ctx := context.Background()
	sess := &mockable.ExperimentSession{
		MockableLogger: log.Log,
	}
	callbacks := handler.NewPrinterCallbacks(sess.Logger())
	err := m.Run(ctx, sess, new(model.Measurement), callbacks)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationFailure(t *testing.T) {
	m := example.NewExperimentMeasurer(example.Config{
		SleepTime:   int64(2 * time.Second),
		ReturnError: true,
	}, "example")
	ctx := context.Background()
	sess := &mockable.ExperimentSession{
		MockableLogger: log.Log,
	}
	callbacks := handler.NewPrinterCallbacks(sess.Logger())
	err := m.Run(ctx, sess, new(model.Measurement), callbacks)
	if err == nil {
		t.Fatal("expected an error here")
	}
}
