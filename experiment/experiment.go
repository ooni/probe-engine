// Package experiment contains network experiment.
package experiment

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/ooni/probe-engine/collector"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const dateFormat = "2006-01-02 15:04:05"

func formatTimeNowUTC() string {
	return time.Now().UTC().Format(dateFormat)
}

// MeasureFunc is the function that fills a measurement.
type MeasureFunc func(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks,
) error

// Experiment is a network experiment.
type Experiment struct {
	// DoMeasure fills a measurement.
	DoMeasure MeasureFunc

	// Callbacks handles experiment events.
	Callbacks handler.Callbacks

	// IncludeProbeIP indicates whether to include the probe IP
	// when submitting measurements.
	IncludeProbeIP bool

	// Report is the report used by this experiment.
	Report *collector.Report

	// Session is the session to which this experiment belongs.
	Session *session.Session

	// TestName is the experiment name.
	TestName string

	// TestStartTime is the UTC time when the test started.
	TestStartTime string

	// TestVersion is the experiment version.
	TestVersion string
}

// New creates a new experiment. You should not call this function directly
// rather you should do <package>.NewExperiment.
func New(
	sess *session.Session, testName, testVersion string, measure MeasureFunc,
) *Experiment {
	return &Experiment{
		DoMeasure:     measure,
		Callbacks:     handler.NewPrinterCallbacks(sess.Logger),
		Session:       sess,
		TestName:      testName,
		TestStartTime: formatTimeNowUTC(),
		TestVersion:   testVersion,
	}
}

// OpenReport opens a new report for the experiment.
func (e *Experiment) OpenReport(ctx context.Context) (err error) {
	if e.Report != nil {
		return // already open
	}
	for _, c := range e.Session.AvailableCollectors {
		if c.Type != "https" {
			e.Session.Logger.Debugf(
				"experiment: unsupported collector type: %s", c.Type,
			)
			continue
		}
		client := &collector.Client{
			BaseURL:    c.Address,
			HTTPClient: e.Session.HTTPDefaultClient, // proxy is OK
			Logger:     e.Session.Logger,
			UserAgent:  e.Session.UserAgent(),
		}
		template := collector.ReportTemplate{
			ProbeASN:        e.Session.ProbeASNString(),
			ProbeCC:         e.Session.ProbeCC(),
			SoftwareName:    e.Session.SoftwareName,
			SoftwareVersion: e.Session.SoftwareVersion,
			TestName:        e.TestName,
			TestVersion:     e.TestVersion,
		}
		e.Report, err = client.OpenReport(ctx, template)
		if err == nil {
			return
		}
		e.Session.Logger.Debugf("experiment: collector error: %s", err.Error())
	}
	err = errors.New("All collectors failed")
	return
}

// ReportID returns the report ID or an empty string, if not open.
func (e *Experiment) ReportID() string {
	if e.Report == nil {
		return ""
	}
	return e.Report.ID
}

func (e *Experiment) newMeasurement(input string) model.Measurement {
	return model.Measurement{
		DataFormatVersion:    "0.2.0",
		Input:                input,
		MeasurementStartTime: formatTimeNowUTC(),
		ProbeIP:              e.Session.ProbeIP(),
		ProbeASN:             e.Session.ProbeASNString(),
		ProbeCC:              e.Session.ProbeCC(),
		ReportID:             e.ReportID(),
		SoftwareName:         e.Session.SoftwareName,
		SoftwareVersion:      e.Session.SoftwareVersion,
		TestName:             e.TestName,
		TestStartTime:        e.TestStartTime,
		TestVersion:          e.TestVersion,
	}
}

// Measure performs a measurement with the specified input.
func (e *Experiment) Measure(
	ctx context.Context, input string,
) (measurement model.Measurement, err error) {
	measurement = e.newMeasurement(input)
	err = e.DoMeasure(ctx, e.Session, &measurement, e.Callbacks)
	return
}

// SubmitMeasurement submits a measurement to the selected collector. It is
// safe to call this function from different goroutines concurrently as long
// as the measurement is not shared by the goroutines.
func (e *Experiment) SubmitMeasurement(
	ctx context.Context, measurement *model.Measurement,
) (err error) {
	if e.Report != nil {
		err = e.Report.SubmitMeasurement(ctx, measurement, e.IncludeProbeIP)
	}
	return
}

// SaveMeasurement saves a measurement on the specified file.
func (e *Experiment) SaveMeasurement(
	measurement model.Measurement, filePath string,
) error {
	data, err := json.Marshal(measurement)
	if err != nil {
		return err
	}
	data = append(data, byte('\n'))
	filep, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := filep.Write(data); err != nil {
		return err
	}
	return filep.Close()
}

// CloseReport closes the open report.
func (e *Experiment) CloseReport(ctx context.Context) (err error) {
	if e.Report != nil {
		err = e.Report.Close(ctx)
		e.Report = nil
	}
	return
}
