package psiphon

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/orchestra/testorchestra"
	"github.com/ooni/probe-engine/model"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestUnitNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "psiphon" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.3.2" {
		t.Fatal("unexpected version")
	}
}

func TestUnitMeasureWithCancelledContext(t *testing.T) {
	m := &measurer{config: makeconfig()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	err := m.Run(
		ctx, newsession(),
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "context canceled") {
		t.Fatal("not the error we expected")
	}
}

func TestUnitMakeWorkingDirEmptyWorkingDir(t *testing.T) {
	r := newRunner(makeconfig(), handler.NewPrinterCallbacks(log.Log), time.Now())
	r.config.WorkDir = ""
	workdir, err := r.makeworkingdir()
	if err == nil {
		t.Fatal("expected an error here")
	}
	if workdir != "" {
		t.Fatal("expected an empty string here")
	}
}

func TestUnitMakeWorkingDirOsRemoveAllError(t *testing.T) {
	r := newRunner(makeconfig(), handler.NewPrinterCallbacks(log.Log), time.Now())
	expected := errors.New("mocked error")
	r.osRemoveAll = func(path string) error {
		return expected
	}
	workdir, err := r.makeworkingdir()
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if workdir != "" {
		t.Fatal("expected an empty string here")
	}
}

func TestUnitMakeWorkingDirOsMkdirAllError(t *testing.T) {
	r := newRunner(makeconfig(), handler.NewPrinterCallbacks(log.Log), time.Now())
	expected := errors.New("mocked error")
	r.osMkdirAll = func(path string, perm os.FileMode) error {
		return expected
	}
	workdir, err := r.makeworkingdir()
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if workdir != "" {
		t.Fatal("expected an empty string here")
	}
}

func TestUnitRunFetchPsiphonConfigError(t *testing.T) {
	r := newRunner(makeconfig(), handler.NewPrinterCallbacks(log.Log), time.Now())
	expected := errors.New("mocked error")
	err := r.run(context.Background(), log.Log, func(context.Context) ([]byte, error) {
		return nil, expected
	})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunMakeworkingdirError(t *testing.T) {
	r := newRunner(makeconfig(), handler.NewPrinterCallbacks(log.Log), time.Now())
	expected := errors.New("mocked error")
	r.osRemoveAll = func(path string) error {
		return expected
	}
	clnt, err := newclient()
	if err != nil {
		t.Fatal(err)
	}
	err = r.run(context.Background(), log.Log, clnt.FetchPsiphonConfig)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunStartTunnelError(t *testing.T) {
	r := newRunner(makeconfig(), handler.NewPrinterCallbacks(log.Log), time.Now())
	err := r.run(context.Background(), log.Log, func(context.Context) ([]byte, error) {
		return []byte("{"), nil
	})
	if !strings.HasSuffix(err.Error(), "unexpected end of JSON input") {
		t.Fatal("not the error we expected")
	}
}

func newclient() (*orchestra.Client, error) {
	clnt := orchestra.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
		statefile.New(kvstore.NewMemoryKeyValueStore()),
	)
	clnt.OrchestrateBaseURL = "https://ps-test.ooni.io"
	clnt.RegistryBaseURL = "https://ps-test.ooni.io"
	ctx := context.Background()
	meta := testorchestra.MetadataFixture()
	if err := clnt.MaybeRegister(ctx, meta); err != nil {
		return nil, err
	}
	if err := clnt.MaybeLogin(ctx); err != nil {
		return nil, err
	}
	return clnt, nil
}

func TestIntegration(t *testing.T) {
	m := &measurer{config: makeconfig()}
	if err := m.Run(
		context.Background(),
		newsession(),
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	); err != nil {
		t.Fatal(err)
	}
}

func TestUnitUsetunnel(t *testing.T) {
	r := newRunner(makeconfig(), handler.NewPrinterCallbacks(log.Log), time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so should fail immediately
	err := r.usetunnel(ctx, 8080, log.Log)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
}

func makeconfig() Config {
	return Config{
		WorkDir: "../../testdata/psiphon_unit_tests",
	}
}

func newsession() model.ExperimentSession {
	return &mockable.ExperimentSession{MockableLogger: log.Log}
}
