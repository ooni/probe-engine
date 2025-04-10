package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/probeservices"
	"github.com/ooni/probe-engine/pkg/registry"
)

func TestCreateAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()

	// Since https://github.com/ooni/probe-cli/pull/1355, some experiments are disabled
	// by default and we need an environment variable to instantiate them
	os.Setenv(registry.OONI_FORCE_ENABLE_EXPERIMENT, "1")
	defer os.Unsetenv(registry.OONI_FORCE_ENABLE_EXPERIMENT)

	for _, name := range AllExperiments() {
		builder, err := sess.NewExperimentBuilder(name)
		if err != nil {
			t.Fatal(err)
		}
		exp := builder.NewExperiment()
		good := (exp.Name() == name)
		if !good {
			// We have introduced the concept of versioned experiments in
			// https://github.com/ooni/probe-cli/pull/882. This works like
			// in brew: we append @vX.Y to the experiment name. So, here
			// we're stripping the version specification and retry.
			index := strings.Index(name, "@")
			if index >= 0 {
				name = name[:index]
				if good := (exp.Name() == name); good {
					continue
				}
			}
			t.Fatal("unexpected experiment name", exp.Name(), name)
		}
	}
}

func TestRunDASH(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
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
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunNdt7(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
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
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("psiphon")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunSNIBlocking(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("sni_blocking")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "kernel.org")
}

func TestRunTelegram(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("telegram")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestRunTor(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("tor")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func TestNeedsInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("web_connectivity")
	if err != nil {
		t.Fatal(err)
	}
	if builder.InputPolicy() != model.InputOrQueryBackend {
		t.Fatal("web_connectivity certainly needs input")
	}
}

func TestSetCallbacks(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.SetOptionAny("SleepTime", 0); err != nil {
		t.Fatal(err)
	}
	register := &registerCallbacksCalled{}
	builder.SetCallbacks(register)
	target := model.NewOOAPIURLInfoWithDefaultCategoryAndCountry("")
	if _, err := builder.NewExperiment().MeasureWithContext(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	if register.onProgressCalled == false {
		t.Fatal("OnProgress not called")
	}
}

type registerCallbacksCalled struct {
	onProgressCalled bool
}

func (c *registerCallbacksCalled) OnProgress(percentage float64, message string) {
	c.onProgressCalled = true
}

func TestCreateInvalidExperiment(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
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
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.SetOptionAny("ReturnError", true); err != nil {
		t.Fatal(err)
	}
	target := model.NewOOAPIURLInfoWithDefaultCategoryAndCountry("")
	measurement, err := builder.NewExperiment().MeasureWithContext(context.Background(), target)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if err.Error() != "mocked error" {
		t.Fatal("unexpected error type")
	}
	if measurement != nil {
		t.Fatal("expected nil measurement here")
	}
}

func TestRunHHFM(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("http_header_field_manipulation")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.NewExperiment(), "")
}

func runexperimentflow(t *testing.T, experiment model.Experiment, input string) {
	ctx := context.Background()
	err := experiment.OpenReportContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if experiment.ReportID() == "" {
		t.Fatal("reportID should not be empty here")
	}
	target := model.NewOOAPIURLInfoWithDefaultCategoryAndCountry(input)
	measurement, err := experiment.MeasureWithContext(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	measurement.AddAnnotations(map[string]string{
		"probe-engine-ci": "yes",
	})
	savedMeasurement, err := json.Marshal(measurement)
	if err != nil {
		t.Fatal(err)
	}
	if savedMeasurement == nil {
		t.Fatal("data is nil")
	}
	tempfile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	filename := tempfile.Name()
	tempfile.Close()
	_, err = experiment.SubmitAndUpdateMeasurementContext(ctx, measurement)
	if err != nil {
		t.Fatal(err)
	}
	err = SaveMeasurement(measurement, filename)
	if err != nil {
		t.Fatal(err)
	}
	if experiment.KibiBytesSent() <= 0 {
		t.Fatal("no data sent?!")
	}
	if experiment.KibiBytesReceived() <= 0 {
		t.Fatal("no data received?!")
	}
	sk := MeasurementSummaryKeys(measurement)
	if sk == nil {
		t.Fatal("got nil summary keys")
	}
	loadedMeasurement, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	withFinalNewline := append(savedMeasurement, '\n')
	if diff := cmp.Diff(withFinalNewline, loadedMeasurement); diff != "" {
		t.Fatal(diff)
	}
}

func TestOpenReportIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
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
	ctx := context.Background()
	if _, err := exp.SubmitAndUpdateMeasurementContext(ctx, &model.Measurement{}); err == nil {
		t.Fatal("we should not be able to submit before OpenReport")
	}
	err = exp.OpenReportContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rid := exp.ReportID()
	if rid == "" {
		t.Fatal("invalid report ID")
	}
	err = exp.OpenReportContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if rid != exp.ReportID() {
		t.Fatal("OpenReport is not idempotent")
	}
}

func TestOpenReportFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		},
	))
	defer server.Close()
	sess := newSessionForTestingNoBackendsLookup(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	exp := builder.NewExperiment().(*experiment)
	exp.session.selectedProbeService = &model.OOAPIService{
		Address: server.URL,
		Type:    "https",
	}
	err = exp.OpenReportContext(context.Background())
	if !strings.HasPrefix(err.Error(), "httpx: request failed") {
		t.Fatal("not the error we expected")
	}
}

func TestOpenReportNewClientFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTestingNoBackendsLookup(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	exp := builder.NewExperiment().(*experiment)
	exp.session.selectedProbeService = &model.OOAPIService{
		Address: "antani:///",
		Type:    "antani",
	}
	err = exp.OpenReportContext(context.Background())
	if !errors.Is(err, probeservices.ErrUnsupportedServiceType) {
		t.Fatal(err)
	}
}

func TestSubmitAndUpdateMeasurementWithClosedReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	builder, err := sess.NewExperimentBuilder("example")
	if err != nil {
		t.Fatal(err)
	}
	exp := builder.NewExperiment()
	m := new(model.Measurement)
	_, err = exp.SubmitAndUpdateMeasurementContext(context.Background(), m)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestMeasureLookupLocationFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	exp := newExperiment(sess, new(antaniMeasurer))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	target := model.NewOOAPIURLInfoWithDefaultCategoryAndCountry("xx")
	if _, err := exp.MeasureWithContext(ctx, target); err == nil {
		t.Fatal("expected an error here")
	}
}

func TestOpenReportNonHTTPS(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	sess.availableProbeServices = []model.OOAPIService{
		{
			Address: "antani",
			Type:    "mascetti",
		},
	}
	exp := newExperiment(sess, new(antaniMeasurer))
	if err := exp.OpenReportContext(context.Background()); err == nil {
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

func (am *antaniMeasurer) Run(ctx context.Context, args *model.ExperimentArgs) error {
	return nil
}
