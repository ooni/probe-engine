package mkrunner_test

import (
	"testing"

	apexlog "github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/mkrunner"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

func TestIntegrationSuccess(t *testing.T) {
	err := mkrunner.Do(
		measurementkit.Settings{},
		session.New(
			apexlog.Log, "ooniprobe-engine", "0.1.0",
			"../../testdata", nil, nil, "../../testdata",
		),
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
		session.New(
			apexlog.Log, "ooniprobe-engine", "0.1.0",
			"../../testdata", nil, nil, "../../testdata",
		),
		&model.Measurement{},
		handler.NewPrinterCallbacks(apexlog.Log),
		mkrunner.FailingStartEx,
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
}
