package engine

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ooni/probe-engine/model"
)

func TestSubmitterNotEnabled(t *testing.T) {
	ctx := context.Background()
	submitter, err := NewSubmitter(ctx, NewSubmitterConfig{
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := submitter.(stubSubmitter); !ok {
		t.Fatal("we did not get a stubSubmitter instance")
	}
	m := new(model.Measurement)
	if err := submitter.SubmitAndUpdateMeasurementContext(ctx, m); err != nil {
		t.Fatal(err)
	}
}

type FakeSubmitterExperiment struct {
	FakeReportID  string
	OpenReportErr error
	SubmitErr     error
}

func (fse FakeSubmitterExperiment) OpenReportContext(ctx context.Context) error {
	return fse.OpenReportErr
}

func (fse FakeSubmitterExperiment) ReportID() string {
	return fse.FakeReportID
}

func (fse FakeSubmitterExperiment) SubmitAndUpdateMeasurementContext(
	ctx context.Context, m *model.Measurement) error {
	if fse.SubmitErr != nil {
		return fse.SubmitErr
	}
	m.ReportID = fse.FakeReportID
	return nil
}

var _ SubmitterExperiment = FakeSubmitterExperiment{}

func TestNewSubmitterOpenReportFailure(t *testing.T) {
	expected := errors.New("mocked error")
	ctx := context.Background()
	submitter, err := NewSubmitter(ctx, NewSubmitterConfig{
		Enabled:    true,
		Experiment: FakeSubmitterExperiment{OpenReportErr: expected},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if submitter != nil {
		t.Fatal("expected nil submitter here")
	}
}

type FakeSubmitterLogger struct {
	Written []string
}

func (fsl *FakeSubmitterLogger) Infof(format string, v ...interface{}) {
	fsl.Written = append(fsl.Written, fmt.Sprintf(format, v...))
}

var _ SubmitterLogger = &FakeSubmitterLogger{}

func TestNewSubmitterOpenReportSuccess(t *testing.T) {
	fakeLogger := &FakeSubmitterLogger{}
	reportID := "a_fake_report_id"
	expected := errors.New("mocked error")
	ctx := context.Background()
	submitter, err := NewSubmitter(ctx, NewSubmitterConfig{
		Enabled: true,
		Experiment: FakeSubmitterExperiment{
			FakeReportID: reportID,
			SubmitErr:    expected,
		},
		Logger: fakeLogger,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := submitter.(realSubmitter); !ok {
		t.Fatal("not the type of submitter we expected")
	}
	if len(fakeLogger.Written) != 1 {
		t.Fatal("written wrong number of log entries")
	}
	if fakeLogger.Written[0] != "reportID: a_fake_report_id" {
		t.Fatal("unexpected lopg entry written")
	}
	m := new(model.Measurement)
	if err := submitter.SubmitAndUpdateMeasurementContext(ctx, m); !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if len(fakeLogger.Written) != 2 {
		t.Fatal("written wrong number of log entries")
	}
	if fakeLogger.Written[1] != "submitting measurement to OONI collector; please, be patient..." {
		t.Fatal("unexpected lopg entry written")
	}
}
