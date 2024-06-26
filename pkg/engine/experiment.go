package engine

//
// Experiment definition and implementation.
//

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"time"

	"github.com/ooni/probe-engine/pkg/bytecounter"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/probeservices"
	"github.com/ooni/probe-engine/pkg/runtimex"
	"github.com/ooni/probe-engine/pkg/version"
)

// experiment implements Experiment.
type experiment struct {
	byteCounter   *bytecounter.Counter
	callbacks     model.ExperimentCallbacks
	measurer      model.ExperimentMeasurer
	report        probeservices.ReportChannel
	session       *Session
	testName      string
	testStartTime string
	testVersion   string
}

// newExperiment creates a new experiment given a measurer.
func newExperiment(sess *Session, measurer model.ExperimentMeasurer) *experiment {
	return &experiment{
		byteCounter:   bytecounter.New(),
		callbacks:     model.NewPrinterCallbacks(sess.Logger()),
		measurer:      measurer,
		session:       sess,
		testName:      measurer.ExperimentName(),
		testStartTime: model.MeasurementFormatTimeNowUTC(),
		testVersion:   measurer.ExperimentVersion(),
	}
}

// KibiBytesReceived implements Experiment.KibiBytesReceived.
func (e *experiment) KibiBytesReceived() float64 {
	return e.byteCounter.KibiBytesReceived()
}

// KibiBytesSent implements Experiment.KibiBytesSent.
func (e *experiment) KibiBytesSent() float64 {
	return e.byteCounter.KibiBytesSent()
}

// Name implements Experiment.Name.
func (e *experiment) Name() string {
	return e.testName
}

// ExperimentMeasurementSummaryKeysNotImplemented is the [model.MeasurementSummary] we use when
// the experiment TestKeys do not provide an implementation of [model.MeasurementSummary].
type ExperimentMeasurementSummaryKeysNotImplemented struct{}

var _ model.MeasurementSummaryKeys = &ExperimentMeasurementSummaryKeysNotImplemented{}

// IsAnomaly implements MeasurementSummary.
func (*ExperimentMeasurementSummaryKeysNotImplemented) Anomaly() bool {
	return false
}

// MeasurementSummaryKeys returns the [model.MeasurementSummaryKeys] associated with a given measurement.
func MeasurementSummaryKeys(m *model.Measurement) model.MeasurementSummaryKeys {
	if tk, ok := m.TestKeys.(model.MeasurementSummaryKeysProvider); ok {
		return tk.MeasurementSummaryKeys()
	}
	return &ExperimentMeasurementSummaryKeysNotImplemented{}
}

// ReportID implements Experiment.ReportID.
func (e *experiment) ReportID() string {
	if e.report == nil {
		return ""
	}
	return e.report.ReportID()
}

// experimentAsyncWrapper makes a sync experiment behave like it was async
type experimentAsyncWrapper struct {
	*experiment
}

var _ model.ExperimentMeasurerAsync = &experimentAsyncWrapper{}

// RunAsync implements ExperimentMeasurerAsync.RunAsync.
func (eaw *experimentAsyncWrapper) RunAsync(
	ctx context.Context, sess model.ExperimentSession, input string,
	callbacks model.ExperimentCallbacks) (<-chan *model.ExperimentAsyncTestKeys, error) {
	out := make(chan *model.ExperimentAsyncTestKeys)
	measurement := eaw.experiment.newMeasurement(input)
	start := time.Now()
	args := &model.ExperimentArgs{
		Callbacks:   eaw.callbacks,
		Measurement: measurement,
		Session:     eaw.session,
	}
	err := eaw.experiment.measurer.Run(ctx, args)
	stop := time.Now()
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(out) // signal the reader we're done!
		out <- &model.ExperimentAsyncTestKeys{
			Extensions:         measurement.Extensions,
			Input:              measurement.Input,
			MeasurementRuntime: stop.Sub(start).Seconds(),
			TestKeys:           measurement.TestKeys,
			TestHelpers:        measurement.TestHelpers,
		}
	}()
	return out, nil
}

// MeasureAsync implements Experiment.MeasureAsync.
func (e *experiment) MeasureAsync(
	ctx context.Context, input string) (<-chan *model.Measurement, error) {
	err := e.session.MaybeLookupLocationContext(ctx) // this already tracks session bytes
	if err != nil {
		return nil, err
	}
	ctx = bytecounter.WithSessionByteCounter(ctx, e.session.byteCounter)
	ctx = bytecounter.WithExperimentByteCounter(ctx, e.byteCounter)
	var async model.ExperimentMeasurerAsync
	if v, okay := e.measurer.(model.ExperimentMeasurerAsync); okay {
		async = v
	} else {
		async = &experimentAsyncWrapper{e}
	}
	in, err := async.RunAsync(ctx, e.session, input, e.callbacks)
	if err != nil {
		return nil, err
	}
	out := make(chan *model.Measurement)
	go func() {
		defer close(out) // we need to signal the consumer we're done
		for tk := range in {
			measurement := e.newMeasurement(input)
			measurement.Extensions = tk.Extensions
			measurement.Input = tk.Input
			measurement.MeasurementRuntime = tk.MeasurementRuntime
			measurement.TestHelpers = tk.TestHelpers
			measurement.TestKeys = tk.TestKeys
			if err := model.ScrubMeasurement(measurement, e.session.ProbeIP()); err != nil {
				// If we fail to scrub the measurement then we are not going to
				// submit it. Most likely causes of error here are unlikely,
				// e.g., the TestKeys being not serializable.
				e.session.Logger().Warnf("can't scrub measurement: %s", err.Error())
				continue
			}
			out <- measurement
		}
	}()
	return out, nil
}

// MeasureWithContext implements Experiment.MeasureWithContext.
func (e *experiment) MeasureWithContext(
	ctx context.Context, input string,
) (measurement *model.Measurement, err error) {
	out, err := e.MeasureAsync(ctx, input)
	if err != nil {
		return nil, err
	}
	for m := range out {
		if measurement == nil {
			measurement = m // as documented just return the first one
		}
	}
	if measurement == nil {
		err = errors.New("experiment returned no measurements")
	}
	return
}

// SubmitAndUpdateMeasurementContext implements Experiment.SubmitAndUpdateMeasurementContext.
func (e *experiment) SubmitAndUpdateMeasurementContext(
	ctx context.Context, measurement *model.Measurement) error {
	if e.report == nil {
		return errors.New("report is not open")
	}
	return e.report.SubmitMeasurement(ctx, measurement)
}

// newMeasurement creates a new measurement for this experiment with the given input.
func (e *experiment) newMeasurement(input string) *model.Measurement {
	utctimenow := time.Now().UTC()
	m := &model.Measurement{
		DataFormatVersion:         model.OOAPIReportDefaultDataFormatVersion,
		Input:                     model.MeasurementTarget(input),
		MeasurementStartTime:      utctimenow.Format(model.MeasurementDateFormat),
		MeasurementStartTimeSaved: utctimenow,
		ProbeIP:                   model.DefaultProbeIP,
		ProbeASN:                  e.session.ProbeASNString(),
		ProbeCC:                   e.session.ProbeCC(),
		ProbeNetworkName:          e.session.ProbeNetworkName(),
		ReportID:                  e.ReportID(),
		ResolverASN:               e.session.ResolverASNString(),
		ResolverIP:                e.session.ResolverIP(),
		ResolverNetworkName:       e.session.ResolverNetworkName(),
		SoftwareName:              e.session.SoftwareName(),
		SoftwareVersion:           e.session.SoftwareVersion(),
		TestName:                  e.testName,
		TestStartTime:             e.testStartTime,
		TestVersion:               e.testVersion,
	}
	m.AddAnnotation("architecture", runtime.GOARCH)
	m.AddAnnotation("engine_name", "ooniprobe-engine")
	m.AddAnnotation("engine_version", version.Version)
	m.AddAnnotation("go_version", runtimex.BuildInfo.GoVersion)
	m.AddAnnotation("platform", e.session.Platform())
	m.AddAnnotation("vcs_modified", runtimex.BuildInfo.VcsModified)
	m.AddAnnotation("vcs_revision", runtimex.BuildInfo.VcsRevision)
	m.AddAnnotation("vcs_time", runtimex.BuildInfo.VcsTime)
	m.AddAnnotation("vcs_tool", runtimex.BuildInfo.VcsTool)
	return m
}

// OpenReportContext implements Experiment.OpenReportContext.
func (e *experiment) OpenReportContext(ctx context.Context) error {
	if e.report != nil {
		return nil // already open
	}
	// use custom client to have proper byte accounting
	httpClient := &http.Client{
		Transport: bytecounter.WrapHTTPTransport(
			e.session.network.HTTPTransport(),
			e.byteCounter,
		),
	}
	client, err := e.session.newProbeServicesClient(ctx)
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

func (e *experiment) newReportTemplate() model.OOAPIReportTemplate {
	return model.OOAPIReportTemplate{
		DataFormatVersion: model.OOAPIReportDefaultDataFormatVersion,
		Format:            model.OOAPIReportDefaultFormat,
		ProbeASN:          e.session.ProbeASNString(),
		ProbeCC:           e.session.ProbeCC(),
		SoftwareName:      e.session.SoftwareName(),
		SoftwareVersion:   e.session.SoftwareVersion(),
		TestName:          e.testName,
		TestStartTime:     e.testStartTime,
		TestVersion:       e.testVersion,
	}
}
