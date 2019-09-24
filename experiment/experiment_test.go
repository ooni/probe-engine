package experiment_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	exp, err := newExperiment(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = exp.OpenReport(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer exp.CloseReport(ctx)
	m, err := exp.Measure(ctx, "xx")
	if err != nil {
		t.Fatal(err)
	}
	if err = exp.SubmitMeasurement(ctx, &m); err != nil {
		t.Fatal(err)
	}
	dirname, err := ioutil.TempDir("", "ooniprobe-engine-experiment-tests")
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(dirname, "report.jsonl")
	if err = exp.SaveMeasurement(m, filename); err != nil {
		t.Fatal(err)
	}
}

func TestOpenReportIdempotent(t *testing.T) {
	ctx := context.Background()
	exp, err := newExperiment(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if exp.ReportID() != "" {
		t.Fatal("unexpected initial report ID")
	}
	if err := exp.SubmitMeasurement(ctx, &model.Measurement{}); err == nil {
		t.Fatal("we should not be able to submit before OpenReport")
	}
	err = exp.OpenReport(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer exp.CloseReport(ctx)
	rid := exp.ReportID()
	if rid == "" {
		t.Fatal("invalid report ID")
	}
	err = exp.OpenReport(ctx)
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
	ctx := context.Background()
	exp, err := newExperiment(ctx)
	if err != nil {
		t.Fatal(err)
	}
	exp.Session.AvailableCollectors = []model.Service{
		model.Service{
			Address: server.URL,
			Type:    "https",
		},
	}
	err = exp.OpenReport(ctx)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func newExperiment(ctx context.Context) (*experiment.Experiment, error) {
	sess := session.New(
		log.Log, "ooniprobe-engine", "0.1.0", "../testdata", nil, nil,
	)
	if err := sess.MaybeLookupBackends(ctx); err != nil {
		return nil, err
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		return nil, err
	}
	return experiment.New(
		sess, "antani", "0.1.1",
		func(
			ctx context.Context,
			sess *session.Session,
			measurement *model.Measurement,
			callbacks handler.Callbacks,
		) error {
			return nil
		}), nil
}
