package example_test

import (
	"context"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

func TestIntegrationSuccess(t *testing.T) {
	m := example.NewExperimentMeasurer(example.Config{
		SleepTime: int64(2 * time.Second),
	}, "example")
	ctx := context.Background()
	sess := &session.Session{
		Logger: log.Log,
	}
	callbacks := handler.NewPrinterCallbacks(sess.Logger)
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
	sess := &session.Session{
		Logger: log.Log,
	}
	callbacks := handler.NewPrinterCallbacks(sess.Logger)
	err := m.Run(ctx, sess, new(model.Measurement), callbacks)
	if err == nil {
		t.Fatal("expected an error here")
	}
}
