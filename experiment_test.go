package engine

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/psiphon"
)

func TestCreateAll(t *testing.T) {
	sess := newSessionForTesting(t)
	for _, name := range AllExperiments() {
		builder, err := sess.NewExperimentBuilder(name)
		if err != nil {
			t.Fatal(err)
		}
		exp := builder.Build()
		if exp.Name() != name {
			t.Fatal("unexpected experiment name")
		}
	}
}

func TestRunExample(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.Build())
}

func runexperimentflow(t *testing.T, experiment *Experiment) {
	err := experiment.OpenReport()
	if err != nil {
		t.Fatal(err)
	}
	measurement, err := experiment.Measure("")
	if err != nil {
		if err == psiphon.ErrDisabled {
			defer experiment.CloseReport()
			return
		}
		t.Fatal(err)
	}
	err = experiment.SubmitAndUpdateMeasurement(measurement)
	if err != nil {
		t.Fatal(err)
	}
	err = experiment.SaveMeasurement(measurement, "/tmp/experiment.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	err = experiment.CloseReport()
	if err != nil {
		t.Fatal(err)
	}
}
