package oonimkall

import (
	"context"
	"encoding/json"
	"time"

	engine "github.com/ooni/probe-engine"
	"github.com/ooni/probe-engine/internal/runtimex"
)

const (
	failureIPLookup              = "failure.ip_lookup"
	failureASNLookup             = "failure.asn_lookup"
	failureCCLookup              = "failure.cc_lookup"
	failureMeasurement           = "failure.measurement"
	failureMeasurementSubmission = "failure.measurement_submission"
	failureReportCreate          = "failure.report_create"
	failureResolverLookup        = "failure.resolver_lookup"
	failureStartup               = "failure.startup"
	measurement                  = "measurement"
	statusEnd                    = "status.end"
	statusGeoIPLookup            = "status.geoip_lookup"
	statusMeasurementDone        = "status.measurement_done"
	statusMeasurementStart       = "status.measurement_start"
	statusMeasurementSubmission  = "status.measurement_submission"
	statusProgress               = "status.progress"
	statusQueued                 = "status.queued"
	statusReportCreate           = "status.report_create"
	statusResolverLookup         = "status.resolver_lookup"
	statusStarted                = "status.started"
)

// runner runs a specific task
type runner struct {
	emitter             *eventEmitter
	maybeLookupLocation func(*engine.Session) error
	out                 chan<- *eventRecord
	settings            *settingsRecord
}

// newRunner creates a new task runner
func newRunner(settings *settingsRecord, out chan<- *eventRecord) *runner {
	return &runner{
		emitter:  newEventEmitter(settings.DisabledEvents, out),
		out:      out,
		settings: settings,
	}
}

func (r *runner) hasUnsupportedSettings(logger *chanLogger) (unsupported bool) {
	sadly := func(why string) {
		r.emitter.EmitFailureStartup(why)
		unsupported = true
	}
	if r.settings.InputFilepaths != nil {
		sadly("InputFilepaths: not supported")
	}
	if r.settings.Options.AllEndpoints != nil {
		sadly("Options.AllEndpoints: not supported")
	}
	if r.settings.Options.Backend != "" {
		sadly("Options.Backend: not supported")
	}
	if r.settings.Options.CABundlePath != "" {
		logger.Warn("Options.CABundlePath: not supported")
	}
	if r.settings.Options.ConstantBitrate != nil {
		logger.Warn("Options.ConstantBitrate: not supported")
	}
	if r.settings.Options.DNSNameserver != nil {
		logger.Warn("Options.DNSNameserver: not supported")
	}
	if r.settings.Options.DNSEngine != nil {
		logger.Warn("Options.DNSEngine: not supported")
	}
	if r.settings.Options.ExpectedBody != nil {
		logger.Warn("Options.ExpectedBody: not supported")
	}
	if r.settings.Options.GeoIPASNPath != "" {
		logger.Warn("Options.GeoIPASNPath: not supported")
	}
	if r.settings.Options.GeoIPCountryPath != "" {
		logger.Warn("Options.GeoIPCountryPath: not supported")
	}
	if r.settings.Options.Hostname != nil {
		logger.Warn("Options.Hostname: not supported")
	}
	if r.settings.Options.IgnoreBouncerError != nil {
		logger.Warn("Options.IgnoreBouncerError: not supported")
	}
	if r.settings.Options.IgnoreOpenReportError != nil {
		logger.Warn("Options.IgnoreOpenReportError: not supported")
	}
	if r.settings.Options.MLabNSAddressFamily != nil {
		logger.Warn("Options.MLabNSAddressFamily: not supported")
	}
	if r.settings.Options.MLabNSBaseURL != nil {
		logger.Warn("Options.MLabNSBaseURL: not supported")
	}
	if r.settings.Options.MLabNSCountry != nil {
		logger.Warn("Options.MLabNSCountry: not supported")
	}
	if r.settings.Options.MLabNSMetro != nil {
		logger.Warn("Options.MLabNSMetro: not supported")
	}
	if r.settings.Options.MLabNSPolicy != nil {
		logger.Warn("Options.MLabNSPolicy: not supported")
	}
	if r.settings.Options.MLabNSToolName != nil {
		logger.Warn("Options.MLabNSToolName: not supported")
	}
	if r.settings.Options.Port != nil {
		sadly("Options.Port: not supported")
	}
	if r.settings.Options.ProbeASN != "" {
		logger.Warn("Options.ProbeASN: not supported")
	}
	if r.settings.Options.ProbeCC != "" {
		logger.Warn("Options.ProbeCC: not supported")
	}
	if r.settings.Options.ProbeIP != "" {
		logger.Warn("Options.ProbeIP: not supported")
	}
	if r.settings.Options.ProbeNetworkName != "" {
		logger.Warn("Options.ProbeNetworkName: not supported")
	}
	if r.settings.Options.RandomizeInput != false {
		sadly("Options.RandomizeInput: not supported")
	}
	if r.settings.Options.SaveRealResolverIP != nil {
		sadly("Options.SaveRealResolverIP: not supported")
	}
	if r.settings.Options.Server != nil {
		sadly("Options.Server: not supported")
	}
	if r.settings.Options.TestSuite != nil {
		sadly("Options.TestSuite: not supported")
	}
	if r.settings.Options.Timeout != nil {
		sadly("Options.Timeout: not supported")
	}
	if r.settings.Options.UUID != nil {
		sadly("Options.UUID: not supported")
	}
	if r.settings.OutputFilepath != "" && r.settings.Options.NoFileReport == false {
		sadly("OutputFilepath && !NoFileReport: not supported")
	}
	// TODO(bassosimone): intercept IgnoreBouncerFailureError and
	// return a failure if such variable is true.
	return
}

func (r *runner) newsession(logger *chanLogger) (*engine.Session, error) {
	kvstore, err := engine.NewFileSystemKVStore(r.settings.StateDir)
	if err != nil {
		return nil, err
	}
	return engine.NewSession(engine.SessionConfig{
		AssetsDir:       r.settings.AssetsDir,
		KVStore:         kvstore,
		Logger:          logger,
		SoftwareName:    r.settings.Options.SoftwareName,
		SoftwareVersion: r.settings.Options.SoftwareVersion,
		TempDir:         r.settings.TempDir,
	})
}

func (r *runner) contextForExperiment(
	ctx context.Context, builder *engine.ExperimentBuilder,
) context.Context {
	if builder.Interruptible() {
		return ctx
	}
	return context.Background()
}

type runnerCallbacks struct {
	emitter *eventEmitter
}

func (cb *runnerCallbacks) OnDataUsage(dloadKiB, uploadKiB float64) {
	// nothing!
}

func (cb *runnerCallbacks) OnProgress(percentage float64, message string) {
	cb.emitter.Emit(statusProgress, eventStatusProgress{
		Percentage: 0.4 + (percentage * 0.6), // open report is 40%
		Message:    message,
	})
}

// Run runs the runner until completion. The context argument controls
// when to stop when processing multiple inputs, as well as when to stop
// experiments explicitly marked as interruptible.
func (r *runner) Run(ctx context.Context) {
	logger := newChanLogger(r.emitter, r.settings.LogLevel, r.out)
	r.emitter.Emit(statusQueued, eventEmpty{})
	if r.hasUnsupportedSettings(logger) {
		return
	}
	r.emitter.Emit(statusStarted, eventEmpty{})
	sess, err := r.newsession(logger)
	if err != nil {
		r.emitter.EmitFailureStartup(err.Error())
		return
	}
	sess.SetIncludeProbeASN(r.settings.Options.SaveRealProbeASN)
	sess.SetIncludeProbeCC(r.settings.Options.SaveRealProbeCC)
	sess.SetIncludeProbeIP(r.settings.Options.SaveRealProbeIP)
	endEvent := new(eventStatusEnd)
	defer func() {
		sess.Close()
		r.emitter.Emit(statusEnd, endEvent)
	}()

	builder, err := sess.NewExperimentBuilder(r.settings.Name)
	if err != nil {
		r.emitter.EmitFailureStartup(err.Error())
		return
	}

	if r.settings.Options.BouncerBaseURL != "" {
		sess.AddAvailableHTTPSBouncer(r.settings.Options.BouncerBaseURL)
	}
	if r.settings.Options.CollectorBaseURL != "" {
		sess.AddAvailableHTTPSCollector(r.settings.Options.CollectorBaseURL)
	}

	if !r.settings.Options.NoBouncer {
		logger.Info("Looking up OONI backends... please, be patient")
		if err := sess.MaybeLookupBackends(); err != nil {
			r.emitter.EmitFailureStartup(err.Error())
			return
		}
		r.emitter.EmitStatusProgress(0.1, "contacted bouncer")
	}
	if !r.settings.Options.NoGeoIP && !r.settings.Options.NoResolverLookup {
		logger.Info("Looking up your location... please, be patient")
		maybeLookupLocation := r.maybeLookupLocation
		if maybeLookupLocation == nil {
			maybeLookupLocation = func(sess *engine.Session) error {
				return sess.MaybeLookupLocation()
			}
		}
		if err := maybeLookupLocation(sess); err != nil {
			r.emitter.EmitFailureGeneric(failureIPLookup, err.Error())
			r.emitter.EmitFailureGeneric(failureASNLookup, err.Error())
			r.emitter.EmitFailureGeneric(failureCCLookup, err.Error())
			r.emitter.EmitFailureGeneric(failureResolverLookup, err.Error())
			return
		}
		r.emitter.EmitStatusProgress(0.2, "geoip lookup")
		r.emitter.EmitStatusProgress(0.3, "resolver lookup")
		r.emitter.Emit(statusGeoIPLookup, eventStatusGeoIPLookup{
			ProbeIP:          sess.ProbeIP(),
			ProbeASN:         sess.ProbeASNString(),
			ProbeCC:          sess.ProbeCC(),
			ProbeNetworkName: sess.ProbeNetworkName(),
		})
		r.emitter.Emit(statusResolverLookup, eventStatusResolverLookup{
			ResolverASN:         sess.ResolverASNString(),
			ResolverIP:          sess.ResolverIP(),
			ResolverNetworkName: sess.ResolverNetworkName(),
		})
	} else if r.settings.Options.NoGeoIP && r.settings.Options.NoResolverLookup {
		logger.Warn("Not looking up your location")
	} else {
		r.emitter.EmitFailureStartup("Inconsistent NoGeoIP and NoResolverLookup options")
		return
	}

	builder.SetCallbacks(&runnerCallbacks{emitter: r.emitter})
	if len(r.settings.Inputs) <= 0 {
		if builder.InputPolicy() == engine.InputRequired {
			r.emitter.EmitFailureStartup("no input provided")
			return
		}
		r.settings.Inputs = append(r.settings.Inputs, "")
	}
	experiment := builder.NewExperiment()
	defer func() {
		endEvent.DownloadedKB = experiment.KibiBytesReceived()
		endEvent.UploadedKB = experiment.KibiBytesSent()
	}()
	if !r.settings.Options.NoCollector {
		logger.Info("Opening report... please, be patient")
		if err := experiment.OpenReport(); err != nil {
			r.emitter.EmitFailureGeneric(failureReportCreate, err.Error())
			return
		}
		defer func() {
			logger.Info("Closing report... please, be patient")
			experiment.CloseReport()
		}()
		r.emitter.EmitStatusProgress(0.4, "open report")
		r.emitter.Emit(statusReportCreate, eventStatusReportGeneric{
			ReportID: experiment.ReportID(),
		})
	}
	// This deviates a little bit from measurement-kit, for which
	// a zero timeout is actually valid. Since it does not make much
	// sense, here we're changing the behaviour.
	//
	// See https://github.com/measurement-kit/measurement-kit/issues/1922
	if r.settings.Options.MaxRuntime > 0 && builder.InputPolicy() == engine.InputRequired {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(
			ctx, time.Duration(r.settings.Options.MaxRuntime)*time.Second,
		)
		defer cancel()
	}
	for idx, input := range r.settings.Inputs {
		if ctx.Err() != nil {
			break
		}
		logger.Infof("Starting measurement with index %d", idx)
		r.emitter.Emit(statusMeasurementStart, eventMeasurementGeneric{
			Idx:   int64(idx),
			Input: input,
		})
		m, err := experiment.MeasureWithContext(
			r.contextForExperiment(ctx, builder),
			input,
		)
		if builder.Interruptible() && ctx.Err() != nil {
			// We want to stop here only if interruptible otherwise we want to
			// submit measurement and stop at beginning of next iteration
			break
		}
		m.AddAnnotations(r.settings.Annotations)
		if err != nil {
			r.emitter.Emit(failureMeasurement, eventMeasurementGeneric{
				Failure: err.Error(),
				Idx:     int64(idx),
				Input:   input,
			})
			// fallthrough: we want to submit the report anyway
		}
		data, err := json.Marshal(m)
		runtimex.PanicOnError(err, "measurement.MarshalJSON failed")
		r.emitter.Emit(measurement, eventMeasurementGeneric{
			Idx:     int64(idx),
			Input:   input,
			JSONStr: string(data),
		})
		if !r.settings.Options.NoCollector {
			logger.Info("Submitting measurement... please, be patient")
			err := experiment.SubmitAndUpdateMeasurement(m)
			r.emitter.Emit(measurementSubmissionEventName(err), eventMeasurementGeneric{
				Idx:     int64(idx),
				Input:   input,
				JSONStr: string(data),
				Failure: measurementSubmissionFailure(err),
			})
		}
		r.emitter.Emit(statusMeasurementDone, eventMeasurementGeneric{
			Idx:   int64(idx),
			Input: input,
		})
	}
}

func measurementSubmissionEventName(err error) string {
	if err != nil {
		return failureMeasurementSubmission
	}
	return statusMeasurementSubmission
}

func measurementSubmissionFailure(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
