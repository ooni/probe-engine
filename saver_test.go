package engine

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/model"
)

func TestNewSaverDisabled(t *testing.T) {
	saver, err := NewSaver(NewSaverConfig{
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := saver.(fakeSaver); !ok {
		t.Fatal("not the type of Saver we expected")
	}
	m := new(model.Measurement)
	if err := saver.SaveMeasurement(m); err != nil {
		t.Fatal(err)
	}
}

func TestNewSaverWithEmptyFilePath(t *testing.T) {
	saver, err := NewSaver(NewSaverConfig{
		Enabled:  true,
		FilePath: "",
	})
	if err == nil || err.Error() != "saver: passed an empty filepath" {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if saver != nil {
		t.Fatal("saver should be nil here")
	}
}

type FakeSaverExperiment struct {
	M        *model.Measurement
	Error    error
	FilePath string
}

func (fse *FakeSaverExperiment) SaveMeasurement(m *model.Measurement, filepath string) error {
	fse.M = m
	fse.FilePath = filepath
	return fse.Error
}

var _ SaverExperiment = &FakeSaverExperiment{}

type FakeSaverLogger struct {
	Written []string
}

func (fsl *FakeSaverLogger) Infof(format string, v ...interface{}) {
	fsl.Written = append(fsl.Written, fmt.Sprintf(format, v...))
}

var _ SaverLogger = &FakeSaverLogger{}

func TestNewSaverWithFailureWhenSaving(t *testing.T) {
	expected := errors.New("mocked error")
	logger := &FakeSaverLogger{}
	fse := &FakeSaverExperiment{Error: expected}
	saver, err := NewSaver(NewSaverConfig{
		Enabled:    true,
		FilePath:   "report.jsonl",
		Experiment: fse,
		Logger:     logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := saver.(realSaver); !ok {
		t.Fatal("not the type of saver we expected")
	}
	m := &model.Measurement{Input: "www.kernel.org"}
	if err := saver.SaveMeasurement(m); !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if len(logger.Written) != 1 {
		t.Fatal("invalid number of log entries")
	}
	if logger.Written[0] != "saving measurement to disk" {
		t.Fatal("invalid logged message")
	}
	if diff := cmp.Diff(fse.M, m); diff != "" {
		t.Fatal(diff)
	}
	if fse.FilePath != "report.jsonl" {
		t.Fatal("passed invalid filepath")
	}
}
