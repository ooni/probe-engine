// Package experiment contains network experiment.
package experiment

import (
	"context"
	"errors"
	"time"

	"github.com/ooni/probe-engine/collector"
	"github.com/ooni/probe-engine/geoiplookup/constants"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

// dateFormat is the format used by OONI for dates inside reports.
const dateFormat = "2006-01-02 15:04:05"

// formatTimeNowUTC formats the current time in UTC using the OONI format.
func formatTimeNowUTC() string {
	return time.Now().UTC().Format(dateFormat)
}

// The Reporter generates the report for an experiment.
type Reporter struct {
	// Report is the report used by this reporter.
	Report *collector.Report

	// Session is the session to which the experiment belongs.
	Session *session.Session

	// TestName is the experiment name.
	TestName string

	// TestStartTime is the UTC time when the test started.
	TestStartTime string

	// TestVersion is the experiment version.
	TestVersion string
}

// New creates a new reporter for an experiment.
func NewReporter(
	session *session.Session, testName, testVersion string,
) *Reporter {
	return &Reporter{
		Session:       session,
		TestName:      testName,
		TestStartTime: formatTimeNowUTC(),
		TestVersion:   testVersion,
	}
}

// OpenReport opens a new report for the experiment.
func (r *Reporter) OpenReport(ctx context.Context) (err error) {
	if r.Report != nil {
		return // already open
	}
	for _, e := range r.Session.AvailableCollectors {
		if e.Type != "https" {
			r.Session.Logger.Debugf(
				"experiment: unsupported collector type: %s", e.Type,
			)
			continue
		}
		client := &collector.Client{
			BaseURL:    e.Address,
			HTTPClient: r.Session.HTTPDefaultClient, // proxy is OK
			Logger:     r.Session.Logger,
			UserAgent:  r.Session.UserAgent(),
		}
		template := collector.ReportTemplate{
			ProbeASN:        r.Session.ProbeASN,
			ProbeCC:         r.Session.ProbeCC,
			SoftwareName:    r.Session.SoftwareName,
			SoftwareVersion: r.Session.SoftwareVersion,
			TestName:        r.TestName,
			TestVersion:     r.TestVersion,
		}
		r.Report, err = client.OpenReport(ctx, template)
		if err == nil {
			return
		}
		r.Session.Logger.Debugf("experiment: collector error: %s", err.Error())
	}
	err = errors.New("All collectors failed")
	return
}

func (r *Reporter) reportID() string {
	if r.Report == nil {
		return ""
	}
	return r.Report.ID
}

// NewMeasurement initializes and returns a new measurement. Note that this
// function will set the ProbeIP to the default, privacy preserving value. If
// you need the ProbeIP when running an experiment and/or you want to submit it
// because the user asked for that, please override its value explicitly.
func (r *Reporter) NewMeasurement(input string) model.Measurement {
	return model.Measurement{
		DataFormatVersion:    "0.2.0",
		Input:                input,
		MeasurementStartTime: formatTimeNowUTC(),
		ProbeIP:              constants.DefaultProbeIP, // privacy by default
		ProbeASN:             r.Session.ProbeASN,
		ProbeCC:              r.Session.ProbeCC,
		ReportID:             r.reportID(),
		SoftwareName:         r.Session.SoftwareName,
		SoftwareVersion:      r.Session.SoftwareVersion,
		TestName:             r.TestName,
		TestStartTime:        r.TestStartTime,
		TestVersion:          r.TestVersion,
	}
}

// SubmitMeasurement submits a measurement to the selected collector. It is
// safe to call this function from different goroutines concurrently as long
// as the measurement is not shared by the goroutines.
func (r *Reporter) SubmitMeasurement(
	ctx context.Context, measurement *model.Measurement,
) (err error) {
	if r.Report != nil {
		err = r.Report.SubmitMeasurement(ctx, measurement)
	}
	return
}

// CloseReport closes the open report.
func (r *Reporter) CloseReport(ctx context.Context) (err error) {
	if r.Report != nil {
		err = r.Report.Close(ctx)
		r.Report = nil
	}
	return
}
