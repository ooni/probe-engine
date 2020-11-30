package engine

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

// TODO(bassosimone): maybe keep track of which measurements
// could not be submitted by a specific submitter?

// Submitter submits a measurement to the OONI collector.
type Submitter interface {
	// SubmitAndUpdateMeasurementContext submits the measurement
	// and updates its report ID field in case of success.
	SubmitAndUpdateMeasurementContext(
		ctx context.Context, m *model.Measurement) error
}

// SubmitterConfig contains settings for NewSubmitter.
type SubmitterConfig struct {
	// Enabled is true if measurement submission is enabled.
	Enabled bool

	// Experiment is the current experiment.
	Experiment SubmitterExperiment

	// Logger is the logger to be used.
	Logger SubmitterLogger
}

// SubmitterExperiment is the Submitter's view of the Experiment.
//
// Implementation note: we don't bother to define a function for closing
// the report here, since closing reports is no longer necessary since
// changes implemented in ooni/api in Oct-Nov 2020.
type SubmitterExperiment interface {
	// ReportID returns the ID of the currently opened report.
	ReportID() string

	// OpenReportContext opens a report for this experiment using the
	// given context to possibly limit the operation duration.
	OpenReportContext(ctx context.Context) error

	// SubmitAndUpdateMeasurementContext submits the measurement
	// and updates its report ID field in case of sucess.
	SubmitAndUpdateMeasurementContext(
		ctx context.Context, m *model.Measurement) error
}

// SubmitterLogger is the logger expected by Submitter.
type SubmitterLogger interface {
	Infof(format string, v ...interface{})
}

// NewSubmitter creates a new submitter instance. Depending on
// whether submission is enabled or not, the returned submitter
// instance migh just be a stub implementation.
func NewSubmitter(ctx context.Context, config SubmitterConfig) (Submitter, error) {
	if !config.Enabled {
		return stubSubmitter{}, nil
	}
	if err := config.Experiment.OpenReportContext(ctx); err != nil {
		return nil, err
	}
	config.Logger.Infof("reportID: %s", config.Experiment.ReportID())
	return realSubmitter{Experiment: config.Experiment, Logger: config.Logger}, nil
}

type stubSubmitter struct{}

func (stubSubmitter) SubmitAndUpdateMeasurementContext(
	ctx context.Context, m *model.Measurement) error {
	return nil
}

var _ Submitter = stubSubmitter{}

type realSubmitter struct {
	Experiment SubmitterExperiment
	Logger     SubmitterLogger
}

func (rs realSubmitter) SubmitAndUpdateMeasurementContext(
	ctx context.Context, m *model.Measurement) error {
	rs.Logger.Infof("submitting measurement to OONI collector; please, be patient...")
	return rs.Experiment.SubmitAndUpdateMeasurementContext(ctx, m)
}

var _ Submitter = realSubmitter{}
