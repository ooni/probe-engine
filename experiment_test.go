package engine

import (
	"errors"
	"testing"

	"github.com/ooni/probe-engine/experiment/example"
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

func TestNeedsInput(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("web_connectivity")
	if err != nil {
		t.Fatal(err)
	}
	if builder.NeedsInput() == false {
		t.Fatal("web_connectivity certainly needs input")
	}
}

func TestSetCallbacks(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.SetOptionInt("SleepTime", 0); err != nil {
		t.Fatal(err)
	}
	register := &registerCallbacksCalled{}
	builder.SetCallbacks(register)
	if _, err := builder.Build().Measure(""); err != nil {
		t.Fatal(err)
	}
	if register.onDataUsageCalled == false {
		t.Fatal("OnDataUsage not called")
	}
	if register.onProgressCalled == false {
		t.Fatal("OnProgress not called")
	}
}

type registerCallbacksCalled struct {
	onProgressCalled  bool
	onDataUsageCalled bool
}

func (c *registerCallbacksCalled) OnDataUsage(dloadKiB, uploadKiB float64) {
	c.onDataUsageCalled = true
}

func (c *registerCallbacksCalled) OnProgress(percentage float64, message string) {
	c.onProgressCalled = true
}

func TestCreateInvalidExperiment(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("antani")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if builder != nil {
		t.Fatal("expected a nil builder here")
	}
}

func TestMeasurementFailure(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.SetOptionBool("ReturnError", true); err != nil {
		t.Fatal(err)
	}
	measurement, err := builder.Build().Measure("")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if err.Error() != "mocked error" {
		t.Fatal("unexpected error type")
	}
	if measurement == nil {
		t.Fatal("expected non nil measurement here")
	}
}

func TestUseOptions(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	options, err := builder.Options()
	if err != nil {
		t.Fatal(err)
	}
	var (
		returnError bool
		message     bool
		sleepTime   bool
		other       int64
	)
	for name, option := range options {
		if name == "ReturnError" {
			returnError = true
			if option.Type != "bool" {
				t.Fatal("ReturnError is not a bool")
			}
			if option.Doc != "Toogle to return a mocked error" {
				t.Fatal("ReturnError doc is wrong")
			}
		} else if name == "Message" {
			message = true
			if option.Type != "string" {
				t.Fatal("Message is not a string")
			}
			if option.Doc != "Message to emit at test completion" {
				t.Fatal("Message doc is wrong")
			}
		} else if name == "SleepTime" {
			sleepTime = true
			if option.Type != "int64" {
				t.Fatal("SleepTime is not an int64")
			}
			if option.Doc != "Amount of time to sleep for" {
				t.Fatal("SleepTime doc is wrong")
			}
		} else {
			other++
		}
	}
	if other != 0 {
		t.Fatal("found unexpected option")
	}
	if !returnError {
		t.Fatal("did not find ReturnError option")
	}
	if !message {
		t.Fatal("did not find Message option")
	}
	if !sleepTime {
		t.Fatal("did not find SleepTime option")
	}
	if err := builder.SetOptionBool("ReturnError", true); err != nil {
		t.Fatal("cannot set ReturnError field")
	}
	if err := builder.SetOptionInt("SleepTime", 10); err != nil {
		t.Fatal("cannot set SleepTime field")
	}
	if err := builder.SetOptionString("Message", "antani"); err != nil {
		t.Fatal("cannot set Message field")
	}
	config := builder.config.(*example.Config)
	if config.ReturnError != true {
		t.Fatal("config.ReturnError was not changed")
	}
	if config.SleepTime != 10 {
		t.Fatal("config.SleepTime was not changed")
	}
	if config.Message != "antani" {
		t.Fatal("config.Message was not changed")
	}
}

func TestRunHHFM(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("http_header_field_manipulation")
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
	if experiment.ReportID() == "" {
		t.Fatal("reportID should not be empty here")
	}
	measurement, err := experiment.Measure("")
	if err != nil {
		if err == psiphon.ErrDisabled {
			defer experiment.CloseReport()
			return
		}
		t.Fatal(err)
	}
	measurement.AddAnnotations(map[string]string{
		"probe-engine-ci": "yes",
	})
	t.Log(measurement.MakeGenericTestKeys())
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

func TestMakeGenericTestKeysMarshalError(t *testing.T) {
	m := new(Measurement)
	out, err := m.makeGenericTestKeys(
		func(interface{}) ([]byte, error) {
			return nil, errors.New("mocked error")
		},
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if out != nil {
		t.Fatal("expected nil output here")
	}
}

func TestOptions(t *testing.T) {
	t.Run("when config is not a pointer", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: 17,
		}
		options, err := b.Options()
		if err == nil {
			t.Fatal("expected an error here")
		}
		if options != nil {
			t.Fatal("expected nil here")
		}
	})
	t.Run("when confg is not a struct", func(t *testing.T) {
		number := 17
		b := &ExperimentBuilder{
			config: &number,
		}
		options, err := b.Options()
		if err == nil {
			t.Fatal("expected an error here")
		}
		if options != nil {
			t.Fatal("expected nil here")
		}
	})
}

func TestSetOption(t *testing.T) {
	t.Run("when config is not a pointer", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: 17,
		}
		if err := b.SetOptionBool("antani", false); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when config is not a struct", func(t *testing.T) {
		number := 17
		b := &ExperimentBuilder{
			config: &number,
		}
		if err := b.SetOptionBool("antani", false); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when field is not valid", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: &ExperimentBuilder{},
		}
		if err := b.SetOptionBool("antani", false); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when field is not bool", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: new(example.Config),
		}
		if err := b.SetOptionBool("Message", false); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when field is not string", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: new(example.Config),
		}
		if err := b.SetOptionString("ReturnError", "xx"); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when field is not int", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: new(example.Config),
		}
		if err := b.SetOptionInt("ReturnError", 17); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when int field does not exist", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: new(example.Config),
		}
		if err := b.SetOptionInt("antani", 17); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when string field does not exist", func(t *testing.T) {
		b := &ExperimentBuilder{
			config: new(example.Config),
		}
		if err := b.SetOptionString("antani", "xx"); err == nil {
			t.Fatal("expected an error here")
		}
	})
}
