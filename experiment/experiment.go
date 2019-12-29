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

// MeasureFunc is the function that performs a measurement.
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
// rather you should do <package>.NewExperiment where <package> is a subpackage
// inside of the experiment package, e.g., `.../experiment/ndt7`.
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

// OpenReport opens a new report for the experiment. This function
// is idempotent.
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
			DataFormatVersion: collector.DefaultDataFormatVersion,
			Format:            collector.DefaultFormat,
			ProbeASN:          e.Session.ProbeASNString(),
			ProbeCC:           e.Session.ProbeCC(),
			SoftwareName:      e.Session.SoftwareName,
			SoftwareVersion:   e.Session.SoftwareVersion,
			TestName:          e.TestName,
			TestVersion:       e.TestVersion,
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
		DataFormatVersion:    collector.DefaultDataFormatVersion,
		Input:                input,
		MeasurementStartTime: formatTimeNowUTC(),
		ProbeIP:              e.Session.ProbeIP(),
		ProbeASN:             e.Session.ProbeASNString(),
		ProbeCC:              e.Session.ProbeCC(),
		ReportID:             e.ReportID(),
		ResolverASN:          e.Session.ResolverASNString(),
		ResolverIP:           e.Session.ResolverIP(),
		ResolverNetworkName:  e.Session.ResolverNetworkName(),
		SoftwareName:         e.Session.SoftwareName,
		SoftwareVersion:      e.Session.SoftwareVersion,
		TestName:             e.TestName,
		TestStartTime:        e.TestStartTime,
		TestVersion:          e.TestVersion,
	}
}

// Measure performs a measurement with the specified input. Note that as
// part of running the measurement, we'll also apply privacy settings. So,
// the measurement you get back is already scrubbed (if needed).
func (e *Experiment) Measure(
	ctx context.Context, input string,
) (measurement model.Measurement, err error) {
	err = e.Session.MaybeLookupLocation(ctx)
	if err != nil {
		return
	}
	measurement = e.newMeasurement(input)
	start := time.Now()
	err = e.DoMeasure(ctx, e.Session, &measurement, e.Callbacks)
	stop := time.Now()
	measurement.MeasurementRuntime = stop.Sub(start).Seconds()
	scrubErr := e.Session.PrivacySettings.Apply(
		&measurement, e.Session.ProbeIP(),
	)
	if err == nil {
		err = scrubErr
	}
	return
}

// SubmitMeasurement submits a measurement to the selected collector. It is
// safe to call this function from different goroutines concurrently as long
// as the measurement is not shared by the goroutines.
func (e *Experiment) SubmitMeasurement(
	ctx context.Context, measurement *model.Measurement,
) error {
	if e.Report == nil {
		return errors.New("Report is not open")
	}
	return e.Report.SubmitMeasurement(ctx, measurement)
}

// SaveMeasurement saves a measurement on the specified file.
func (e *Experiment) SaveMeasurement(
	measurement model.Measurement, filePath string,
) error {
	return e.SaveMeasurementEx(
		measurement, filePath, json.Marshal, os.OpenFile,
		func(fp *os.File, b []byte) (int, error) {
			return fp.Write(b)
		},
	)
}

// SaveMeasurementEx is like SaveMeasurement but allows you to mock
// any operation that SaveMeasurement would perform. You generally
// want to call SaveMeasurement rather than this function.
func (e *Experiment) SaveMeasurementEx(
	measurement model.Measurement, filePath string,
	marshal func(v interface{}) ([]byte, error),
	openFile func(name string, flag int, perm os.FileMode) (*os.File, error),
	write func(fp *os.File, b []byte) (n int, err error),
) error {
	data, err := marshal(measurement)
	if err != nil {
		return err
	}
	data = append(data, byte('\n'))
	filep, err := openFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := write(filep, data); err != nil {
		return err
	}
	return filep.Close()
}

// CloseReport closes the open report. This function is idempotent.
func (e *Experiment) CloseReport(ctx context.Context) (err error) {
	if e.Report != nil {
		err = e.Report.Close(ctx)
		e.Report = nil
	}
	return
}
