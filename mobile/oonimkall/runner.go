package oonimkall

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"github.com/m-lab/go/rtx"
	engine "github.com/ooni/probe-engine"
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
	emitter  *eventEmitter
	out      chan<- *eventRecord
	settings *settingsRecord
}

// newRunner creates a new task runner
func newRunner(settings *settingsRecord, out chan<- *eventRecord) *runner {
	return &runner{
		emitter:  newEventEmitter(settings.DisabledEvents, out),
		out:      out,
		settings: settings,
	}
}

func (r *runner) hasUnsupportedSettings() (unsupported bool) {
	sadly := func(why string) {
		r.emitter.EmitFailureStartup(why)
		unsupported = true
	}
	if r.settings.InputFilepaths != nil {
		sadly("InputFilepaths: not supported")
	}
	if r.settings.Options.Backend != "" {
		sadly("Options.Backend: not supported")
	}
	if r.settings.Options.CaBundlePath != "" {
		sadly("Options.CaBundlePath: not supported")
	}
	if r.settings.Options.GeoIPASNPath != "" {
		sadly("Options.GeoIPASNPath: not supported")
	}
	if r.settings.Options.GeoIPCountryPath != "" {
		sadly("Options.GeoIPCountryPath: not supported")
	}
	if r.settings.Options.NoFileReport != nil {
		sadly("Options.NoFileReport: not supported")
	}
	if r.settings.Options.ProbeASN != "" {
		sadly("Options.ProbeASN: not supported")
	}
	if r.settings.Options.ProbeCC != "" {
		sadly("Options.ProbeCC: not supported")
	}
	if r.settings.Options.ProbeIP != "" {
		sadly("Options.ProbeIP: not supported")
	}
	if r.settings.Options.ProbeNetworkName != "" {
		sadly("Options.ProbeNetworkName: not supported")
	}
	if r.settings.Options.RandomizeInput != nil {
		sadly("Options.RandomizeInput: not supported")
	}
	if r.settings.OutputFilePath != "" {
		sadly("OutputFilePath: not supported")
	}
	// TODO(bassosimone): intercept IgnoreBouncerFailureError and
	// return a failure if such variable is true.
	return
}

// Run runs the runner until completion
func (r *runner) Run(ctx context.Context) {
	// TODO(bassosimone): accurately count bytes
	// TODO(bassosimone): honour context
	// TODO(bassosimone): intercept all options we ignore

	defer r.emitter.Emit(statusEnd, eventValue{})
	r.emitter.Emit(statusQueued, eventValue{})
	if r.hasUnsupportedSettings() {
		return
	}
	r.emitter.Emit(statusStarted, eventValue{})

	logger := newChanLogger(r.emitter, r.settings, r.out)
	tlsconf := new(tls.Config)
	if r.settings.Options.CaBundlePath != "" {
		certdata, err := ioutil.ReadFile(r.settings.Options.CaBundlePath)
		if err != nil {
			r.emitter.EmitFailureStartup(err.Error())
			return
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(certdata) {
			r.emitter.EmitFailureStartup("AppendCertsFromPEM failed")
			return
		}
	}
	if r.settings.StateDir == "" {
		r.emitter.EmitFailureStartup("options.state_dir is emtpy")
		return
	}
	kvstore, err := engine.NewFileSystemKVStore(r.settings.StateDir)
	if err != nil {
		r.emitter.EmitFailureStartup(err.Error())
		return
	}
	sess, err := engine.NewSession(engine.SessionConfig{
		AssetsDir:       r.settings.AssetsDir,
		KVStore:         kvstore,
		Logger:          logger,
		SoftwareName:    r.settings.Options.SoftwareName,
		SoftwareVersion: r.settings.Options.SoftwareVersion,
		TLSConfig:       tlsconf,
		TempDir:         r.settings.TempDir,
	})
	if err != nil {
		r.emitter.EmitFailureStartup(err.Error())
		return
	}

	// TODO(bassosimone):
	//
	// 1. set experiment options here
	//
	// 2. we should probably also set callbacks here?
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
		logger.Info("Looking up OONI backends")
		if err := sess.MaybeLookupBackends(); err != nil {
			r.emitter.EmitFailureStartup(err.Error())
			return
		}
		r.emitter.EmitStatusProgress(0.1, "contacted bouncer")
	}
	if !r.settings.Options.NoGeoIP && !r.settings.Options.NoResolverLookup {
		logger.Info("Looking up your location")
		if err := sess.MaybeLookupLocation(); err != nil {
			r.emitter.EmitFailure(failureIPLookup, err.Error())
			r.emitter.EmitFailure(failureASNLookup, err.Error())
			r.emitter.EmitFailure(failureCCLookup, err.Error())
			r.emitter.EmitFailure(failureResolverLookup, err.Error())
			return
		}
		r.emitter.EmitStatusProgress(0.2, "geoip lookup")
		r.emitter.EmitStatusProgress(0.3, "resolver lookup")
		r.emitter.Emit(statusGeoIPLookup, eventValue{
			ProbeIP:          sess.ProbeIP(),
			ProbeASN:         sess.ProbeASNString(),
			ProbeCC:          sess.ProbeCC(),
			ProbeNetworkName: sess.ProbeNetworkName(),
		})
		r.emitter.Emit(statusResolverLookup, eventValue{
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

	if len(r.settings.Inputs) <= 0 {
		if builder.NeedsInput() {
			r.emitter.EmitFailureStartup("no input provided")
			return
		}
		r.settings.Inputs = append(r.settings.Inputs, "")
	}
	experiment := builder.Build()
	if !r.settings.Options.NoCollector {
		if err := experiment.OpenReport(); err != nil {
			r.emitter.EmitFailure(failureReportCreate, err.Error())
			return
		}
		defer experiment.CloseReport()
		r.emitter.EmitStatusProgress(0.4, "open report")
		r.emitter.Emit(statusReportCreate, eventValue{
			ReportID: experiment.ReportID(),
		})
	}
	if r.settings.Options.MaxRuntime >= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(
			ctx, time.Duration(r.settings.Options.MaxRuntime)*time.Second,
		)
		defer cancel()
	}
	for idx, input := range r.settings.Inputs {
		r.emitter.Emit(statusMeasurementStart, eventValue{
			Idx:   int64(idx),
			Input: input,
		})
		m, err := experiment.Measure(input)
		m.AddAnnotations(r.settings.Annotations)
		if err != nil {
			r.emitter.Emit(failureMeasurement, eventValue{
				Failure: err.Error(),
				Idx:     int64(idx),
				Input:   input,
			})
			// fallthrough: we want to submit the report anyway
		}
		data, err := m.MarshalJSON()
		rtx.PanicOnError(err, "measurement.MarshalJSON failed")
		r.emitter.Emit(measurement, eventValue{
			Idx:     int64(idx),
			Input:   input,
			JSONStr: string(data),
		})
		if !r.settings.Options.NoCollector {
			if err := experiment.SubmitAndUpdateMeasurement(m); err != nil {
				r.emitter.Emit(failureMeasurementSubmission, eventValue{
					Idx:     int64(idx),
					Input:   input,
					JSONStr: string(data),
					Failure: err.Error(),
				})
			} else {
				r.emitter.Emit(statusMeasurementSubmission, eventValue{
					Idx:   int64(idx),
					Input: input,
				})
			}
		}
		r.emitter.Emit(statusMeasurementDone, eventValue{
			Idx:   int64(idx),
			Input: input,
		})
	}
}
