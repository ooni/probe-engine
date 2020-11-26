// Package engine contains the engine API
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ooni/probe-engine/internal/platform"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/probeservices"
	"github.com/ooni/probe-engine/resources"
)

const dateFormat = "2006-01-02 15:04:05"

func formatTimeNowUTC() string {
	return time.Now().UTC().Format(dateFormat)
}

// Experiment is an experiment instance.
type Experiment struct {
	byteCounter   *bytecounter.Counter
	callbacks     model.ExperimentCallbacks
	measurer      model.ExperimentMeasurer
	report        *probeservices.Report
	session       *Session
	testName      string
	testStartTime string
	testVersion   string
}

// NewExperiment creates a new experiment given a measurer. The preferred
// way to create an experiment is the ExperimentBuilder. Though this function
// allows the programmer to create a custom, external experiment.
func NewExperiment(sess *Session, measurer model.ExperimentMeasurer) *Experiment {
	return &Experiment{
		byteCounter:   bytecounter.New(),
		callbacks:     model.NewPrinterCallbacks(sess.Logger()),
		measurer:      measurer,
		session:       sess,
		testName:      measurer.ExperimentName(),
		testStartTime: formatTimeNowUTC(),
		testVersion:   measurer.ExperimentVersion(),
	}
}

// KibiBytesReceived accounts for the KibiBytes received by the HTTP clients
// managed by this session so far, including experiments.
func (e *Experiment) KibiBytesReceived() float64 {
	return e.byteCounter.KibiBytesReceived()
}

// KibiBytesSent is like KibiBytesReceived but for the bytes sent.
func (e *Experiment) KibiBytesSent() float64 {
	return e.byteCounter.KibiBytesSent()
}

// Name returns the experiment name.
func (e *Experiment) Name() string {
	return e.testName
}

// GetSummaryKeys returns a data structure containing a
// summary of the test keys for probe-cli.
func (e *Experiment) GetSummaryKeys(m *model.Measurement) (interface{}, error) {
	return e.measurer.GetSummaryKeys(m)
}

// OpenReport is an idempotent method to open a report. We assume that
// you have configured the available probe services, either manually or
// through using the session's MaybeLookupBackends method.
func (e *Experiment) OpenReport() (err error) {
	return e.openReport(context.Background())
}

// ReportID returns the open reportID, if we have opened a report
// successfully before, or an empty string, otherwise.
func (e *Experiment) ReportID() string {
	if e.report == nil {
		return ""
	}
	return e.report.ID
}

// LoadMeasurement loads a measurement from a byte stream. The measurement
// must be a measurement for this experiment.
func (e *Experiment) LoadMeasurement(data []byte) (*model.Measurement, error) {
	var measurement model.Measurement
	if err := json.Unmarshal(data, &measurement); err != nil {
		return nil, err
	}
	if measurement.TestName != e.Name() {
		return nil, errors.New("not a measurement for this experiment")
	}
	return &measurement, nil
}

// Measure performs a measurement with input. We assume that you have
// configured the available test helpers, either manually or by calling
// the session's MaybeLookupBackends() method.
func (e *Experiment) Measure(input string) (*model.Measurement, error) {
	return e.MeasureWithContext(context.Background(), input)
}

// MeasureWithContext is like Measure but with context.
func (e *Experiment) MeasureWithContext(
	ctx context.Context, input string,
) (measurement *model.Measurement, err error) {
	err = e.session.MaybeLookupLocationContext(ctx) // this already tracks session bytes
	if err != nil {
		return
	}
	ctx = dialer.WithSessionByteCounter(ctx, e.session.byteCounter)
	ctx = dialer.WithExperimentByteCounter(ctx, e.byteCounter)
	measurement = e.newMeasurement(input)
	start := time.Now()
	err = e.measurer.Run(ctx, e.session, measurement, &sessionExperimentCallbacks{
		exp:   e,
		inner: e.callbacks,
		sess:  e.session,
	})
	stop := time.Now()
	measurement.MeasurementRuntime = stop.Sub(start).Seconds()
	scrubErr := measurement.Scrub(e.session.ProbeIP())
	if err == nil {
		err = scrubErr
	}
	return
}

type sessionExperimentCallbacks struct {
	exp   *Experiment
	inner model.ExperimentCallbacks
	sess  *Session
}

func (cb *sessionExperimentCallbacks) OnProgress(percentage float64, message string) {
	cb.inner.OnProgress(percentage, message)
}

// SaveMeasurement saves a measurement on the specified file path.
func (e *Experiment) SaveMeasurement(measurement *model.Measurement, filePath string) error {
	return e.saveMeasurement(
		measurement, filePath, json.Marshal, os.OpenFile,
		func(fp *os.File, b []byte) (int, error) {
			return fp.Write(b)
		},
	)
}

// SubmitAndUpdateMeasurement submits a measurement and updates the
// fields whose value has changed as part of the submission.
func (e *Experiment) SubmitAndUpdateMeasurement(measurement *model.Measurement) error {
	if e.report == nil {
		return errors.New("Report is not open")
	}
	return e.report.SubmitMeasurement(context.Background(), measurement)
}

// CloseReport is an idempotent method that closes an open report
// if one has previously been opened, otherwise it does nothing.
func (e *Experiment) CloseReport() (err error) {
	if e.report != nil {
		err = e.report.Close(context.Background())
		e.report = nil
	}
	return
}

func (e *Experiment) newMeasurement(input string) *model.Measurement {
	utctimenow := time.Now().UTC()
	m := model.Measurement{
		DataFormatVersion:         probeservices.DefaultDataFormatVersion,
		Input:                     model.MeasurementTarget(input),
		MeasurementStartTime:      utctimenow.Format(dateFormat),
		MeasurementStartTimeSaved: utctimenow,
		ProbeIP:                   e.session.MaybeProbeIP(),
		ProbeASN:                  e.session.MaybeProbeASNString(),
		ProbeCC:                   e.session.MaybeProbeCC(),
		ProbeNetworkName:          e.session.MaybeProbeNetworkName(),
		ReportID:                  e.ReportID(),
		ResolverASN:               e.session.MaybeResolverASNString(),
		ResolverIP:                e.session.MaybeResolverIP(),
		ResolverNetworkName:       e.session.MaybeResolverNetworkName(),
		SoftwareName:              e.session.SoftwareName(),
		SoftwareVersion:           e.session.SoftwareVersion(),
		TestName:                  e.testName,
		TestStartTime:             e.testStartTime,
		TestVersion:               e.testVersion,
	}
	m.AddAnnotation("assets_version", strconv.FormatInt(resources.Version, 10))
	m.AddAnnotation("engine_name", "ooniprobe-engine")
	m.AddAnnotation("engine_version", Version)
	m.AddAnnotation("platform", platform.Name())
	return &m
}

func (e *Experiment) openReport(ctx context.Context) error {
	if e.report != nil {
		return nil // already open
	}
	// use custom client to have proper byte accounting
	httpClient := &http.Client{
		Transport: &httptransport.ByteCountingTransport{
			RoundTripper: e.session.httpDefaultTransport, // proxy is OK
			Counter:      e.byteCounter,
		},
	}
	if e.session.selectedProbeService == nil {
		return errors.New("no probe services selected")
	}
	client, err := probeservices.NewClient(e.session, *e.session.selectedProbeService)
	if err != nil {
		e.session.logger.Debugf("%+v", err)
		return err
	}
	client.HTTPClient = httpClient // patch HTTP client to use
	template := e.newReportTemplate()
	e.report, err = client.OpenReport(ctx, template)
	if err != nil {
		e.session.logger.Debugf("experiment: probe services error: %s", err.Error())
		return err
	}
	return nil
}

func (e *Experiment) newReportTemplate() probeservices.ReportTemplate {
	return probeservices.ReportTemplate{
		DataFormatVersion: probeservices.DefaultDataFormatVersion,
		Format:            probeservices.DefaultFormat,
		ProbeASN:          e.session.MaybeProbeASNString(),
		ProbeCC:           e.session.MaybeProbeCC(),
		SoftwareName:      e.session.SoftwareName(),
		SoftwareVersion:   e.session.SoftwareVersion(),
		TestName:          e.testName,
		TestStartTime:     e.testStartTime,
		TestVersion:       e.testVersion,
	}
}

func (e *Experiment) saveMeasurement(
	measurement *model.Measurement, filePath string,
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
