package mkrunner_test

import (
	"testing"

	apexlog "github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/mkrunner"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

func TestIntegrationSuccess(t *testing.T) {
	err := mkrunner.Do(
		measurementkit.Settings{},
		&mockable.ExperimentSession{MockableLogger: apexlog.Log},
		&model.Measurement{},
		handler.NewPrinterCallbacks(apexlog.Log),
		mkrunner.DoNothingStartEx,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationFailure(t *testing.T) {
	err := mkrunner.Do(
		measurementkit.Settings{},
		&mockable.ExperimentSession{MockableLogger: apexlog.Log},
		&model.Measurement{},
		handler.NewPrinterCallbacks(apexlog.Log),
		mkrunner.FailingStartEx,
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
}
