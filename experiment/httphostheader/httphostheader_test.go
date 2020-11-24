package httphostheader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "http_host_header" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.2.0" {
		t.Fatal("unexpected version")
	}
}

func TestMeasurerMeasureNoMeasurementInput(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{
		TestHelperURL: "http://www.google.com",
	})
	err := measurer.Run(
		context.Background(),
		newsession(),
		new(model.Measurement),
		model.NewPrinterCallbacks(log.Log),
	)
	if err == nil || err.Error() != "experiment requires input" {
		t.Fatal("not the error we expected")
	}
}

func TestMeasurerMeasureNoTestHelper(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	measurement := &model.Measurement{Input: "x.org"}
	err := measurer.Run(
		context.Background(),
		newsession(),
		measurement,
		model.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunnerHTTPSetHostHeader(t *testing.T) {
	var host string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host = r.Host
		w.WriteHeader(200)
	}))
	defer server.Close()
	measurer := NewExperimentMeasurer(Config{
		TestHelperURL: server.URL,
	})
	measurement := &model.Measurement{
		Input: "x.org",
	}
	err := measurer.Run(
		context.Background(),
		newsession(),
		measurement,
		model.NewPrinterCallbacks(log.Log),
	)
	if host != "x.org" {
		t.Fatal("not the host we expected")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func newsession() model.ExperimentSession {
	return &mockable.Session{MockableLogger: log.Log}
}
