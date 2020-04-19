package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

func TestCreateAll(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	for _, name := range AllExperiments() {
		builder, err := sess.NewExperimentBuilder(name)
		if err != nil {
			t.Fatal(err)
		}
		exp := builder.NewExperiment()
		good := (exp.Name() == name)
		if !good {
			t.Fatal("unexpected experiment name")
		}
	}
}

func TestRunDASH(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("dash")
	if err != nil {
		t.Fatal(err)
	}
	if !builder.Interruptible() {
		t.Fatal("dash not marked as interruptible")
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunExample(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunNdt7(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("ndt7")
	if err != nil {
		t.Fatal(err)
	}
	if !builder.Interruptible() {
		t.Fatal("ndt7 not marked as interruptible")
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunPsiphon(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("psiphon")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunSNIBlocking(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("sni_blocking")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "kernel.org")
}

func TestRunTelegram(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("telegram")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunTor(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("tor")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestNeedsInput(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
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
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.SetOptionInt("SleepTime", 0); err != nil {
		t.Fatal(err)
	}
	register := &registerCallbacksCalled{}
	builder.SetCallbacks(register)
	if _, err := builder.NewExperiment().Measure(""); err != nil {
		t.Fatal(err)
	}
	if register.onProgressCalled == false {
		t.Fatal("OnProgress not called")
	}
}

type registerCallbacksCalled struct {
	onProgressCalled bool
}

func (c *registerCallbacksCalled) OnDataUsage(dloadKiB, uploadKiB float64) {
	// nothing - unused
}

func (c *registerCallbacksCalled) OnProgress(percentage float64, message string) {
	c.onProgressCalled = true
}

func TestCreateInvalidExperiment(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
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
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.SetOptionBool("ReturnError", true); err != nil {
		t.Fatal(err)
	}
	measurement, err := builder.NewExperiment().Measure("")
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
	defer sess.Close()
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
	if !measurementkit.Available() {
		t.Skip("Measurement Kit not available; skipping")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("http_header_field_manipulation")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func runexperimentflow(t *testing.T, experiment *Experiment, input string) {
	err := experiment.OpenReport()
	if err != nil {
		t.Fatal(err)
	}
	if experiment.ReportID() == "" {
		t.Fatal("reportID should not be empty here")
	}
	measurement, err := experiment.Measure(input)
	if err != nil {
		t.Fatal(err)
	}
	measurement.AddAnnotations(map[string]string{
		"probe-engine-ci": "yes",
	})
	data, err := json.Marshal(measurement)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("data is nil")
	}
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

func TestLoadMeasurement(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	experiment := builder.NewExperiment()
	testflow := func(t *testing.T, name string) (*model.Measurement, error) {
		path := fmt.Sprintf(
			"testdata/loadable-measurement-%s.jsonl", name,
		)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return experiment.LoadMeasurement(data)
	}
	t.Run("with correct name", func(t *testing.T) {
		measurement, err := testflow(t, "example")
		if err != nil {
			t.Fatal(err)
		}
		if measurement == nil {
			t.Fatal("expected non nil measurement here")
		}
	})
	t.Run("with invalid name", func(t *testing.T) {
		measurement, err := testflow(t, "wrongname")
		if err == nil {
			t.Fatal("expected error here")
		}
		if measurement != nil {
			t.Fatal("expected nil measurement here")
		}
		if err.Error() != "not a measurement for this experiment" {
			t.Fatal("unexpected error value")
		}
	})
	t.Run("with invalid JSON", func(t *testing.T) {
		measurement, err := testflow(t, "notjson")
		if err == nil {
			t.Fatal("expected error here")
		}
		if measurement != nil {
			t.Fatal("expected nil measurement here")
		}
		if err.Error() == "not a measurement for this experiment" {
			t.Fatal("unexpected error value")
		}
	})
}

func TestSaveMeasurementErrors(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	exp := builder.NewExperiment()
	dirname, err := ioutil.TempDir("", "ooniprobe-engine-save-measurement")
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(dirname, "report.jsonl")
	m := new(model.Measurement)
	err = exp.SaveMeasurementEx(
		m, filename, func(v interface{}) ([]byte, error) {
			return nil, errors.New("mocked error")
		}, os.OpenFile, func(fp *os.File, b []byte) (int, error) {
			return fp.Write(b)
		},
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	err = exp.SaveMeasurementEx(
		m, filename, json.Marshal,
		func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return nil, errors.New("mocked error")
		}, func(fp *os.File, b []byte) (int, error) {
			return fp.Write(b)
		},
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	err = exp.SaveMeasurementEx(
		m, filename, json.Marshal, os.OpenFile,
		func(fp *os.File, b []byte) (int, error) {
			return 0, errors.New("mocked error")
		},
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestOpenReportIdempotent(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	exp := builder.NewExperiment()
	if exp.ReportID() != "" {
		t.Fatal("unexpected initial report ID")
	}
	if err := exp.SubmitAndUpdateMeasurement(&model.Measurement{}); err == nil {
		t.Fatal("we should not be able to submit before OpenReport")
	}
	err = exp.OpenReport()
	if err != nil {
		t.Fatal(err)
	}
	defer exp.CloseReport()
	rid := exp.ReportID()
	if rid == "" {
		t.Fatal("invalid report ID")
	}
	err = exp.OpenReport()
	if err != nil {
		t.Fatal(err)
	}
	if rid != exp.ReportID() {
		t.Fatal("OpenReport is not idempotent")
	}
}

func TestOpenReportFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		},
	))
	defer server.Close()
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	exp := builder.NewExperiment()
	exp.session.availableCollectors = []model.Service{
		{
			Address: server.URL,
			Type:    "https",
		},
	}
	err = exp.OpenReport()
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestSubmitAndUpdateMeasurementWithClosedReport(t *testing.T) {
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	exp := builder.NewExperiment()
	m := new(model.Measurement)
	err = exp.SubmitAndUpdateMeasurement(m)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestMeasureLookupLocationFailure(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	exp := NewExperiment(sess, new(antaniMeasurer))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	if _, err := exp.MeasureWithContext(ctx, "xx"); err == nil {
		t.Fatal("expected an error here")
	}
}

func TestOpenReportNonHTTPS(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	sess.availableCollectors = []model.Service{
		{
			Address: "antani",
			Type:    "mascetti",
		},
	}
	exp := NewExperiment(sess, new(antaniMeasurer))
	if err := exp.OpenReport(); err == nil {
		t.Fatal("expected an error here")
	}
}

type antaniMeasurer struct{}

func (am *antaniMeasurer) ExperimentName() string {
	return "antani"
}

func (am *antaniMeasurer) ExperimentVersion() string {
	return "0.1.1"
}

func (am *antaniMeasurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	return nil
}
