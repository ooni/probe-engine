package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/model"
)

func TestSubmitterNotEnabled(t *testing.T) {
	ctx := context.Background()
	submitter, err := NewSubmitter(ctx, SubmitterConfig{
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
	submitter, err := NewSubmitter(ctx, SubmitterConfig{
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

func TestNewSubmitterOpenReportSuccess(t *testing.T) {
	reportID := "a_fake_report_id"
	expected := errors.New("mocked error")
	ctx := context.Background()
	submitter, err := NewSubmitter(ctx, SubmitterConfig{
		Enabled: true,
		Experiment: FakeSubmitterExperiment{
			FakeReportID: reportID,
			SubmitErr:    expected,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	m := new(model.Measurement)
	if err := submitter.SubmitAndUpdateMeasurementContext(ctx, m); !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
}
