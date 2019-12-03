// +build !nopsiphon

package psiphon

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestUnitNewExperiment(t *testing.T) {
	sess := session.New(
		log.Log, softwareName, softwareVersion,
		"../../testdata", nil, nil, "../../testdata",
	)
	experiment := NewExperiment(sess, makeconfig())
	if experiment == nil {
		t.Fatal("nil experiment returned")
	}
}

func TestUnitMeasureWithCancelledContext(t *testing.T) {
	m := &measurer{config: makeconfig()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	err := m.measure(
		ctx, &session.Session{Logger: log.Log},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "controller.Run exited unexpectedly") {
		t.Fatal("not the error we expected")
	}
}

func TestUnitReadConfig(t *testing.T) {
	r := newRunner(makeconfig())
	r.config.ConfigFilePath = "../../testdata/nonexistent_config.json"
	configJSON, err := r.readconfig()
	if err == nil {
		t.Fatal("expected an error here")
	}
	if configJSON != nil {
		t.Fatal("expected nil configJSON")
	}
}

func TestUnitMakeWorkingDirEmptyWorkingDir(t *testing.T) {
	r := newRunner(makeconfig())
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
	r := newRunner(makeconfig())
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
	r := newRunner(makeconfig())
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

func TestUnitRunReadconfigError(t *testing.T) {
	r := newRunner(makeconfig())
	r.config.ConfigFilePath = "../../testdata/nonexistent_config.json"
	err := r.run(context.Background(), log.Log)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestUnitRunMakeworkingdirError(t *testing.T) {
	r := newRunner(makeconfig())
	expected := errors.New("mocked error")
	r.osRemoveAll = func(path string) error {
		return expected
	}
	err := r.run(context.Background(), log.Log)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestIntegration(t *testing.T) {
	m := &measurer{config: makeconfig()}
	sess := &session.Session{Logger: log.Log}
	if err := m.measure(
		context.Background(),
		sess,
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	); err != nil {
		t.Fatal(err)
	}
}

func TestUnitUsetunnel(t *testing.T) {
	r := newRunner(makeconfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so should fail immediately
	err := r.usetunnel(ctx, 8080, log.Log)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
}

func makeconfig() Config {
	return Config{
		ConfigFilePath: "../../testdata/psiphon_config.json",
		WorkDir:        "../../testdata/psiphon_unit_tests",
	}
}
