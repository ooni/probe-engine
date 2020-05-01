package urlgetter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestMeasurer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := urlgetter.NewExperimentMeasurer(urlgetter.Config{})
	if m.ExperimentName() != "urlgetter" {
		t.Fatal("invalid experiment name")
	}
	if m.ExperimentVersion() != "0.0.3" {
		t.Fatal("invalid experiment version")
	}
	measurement := new(model.Measurement)
	measurement.Input = "https://www.google.com"
	err := m.Run(
		ctx, &mockable.ExperimentSession{},
		measurement, handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if len(measurement.Extensions) != 4 {
		t.Fatal("not the expected number of extensions")
	}
}
