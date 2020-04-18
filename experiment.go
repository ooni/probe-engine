// Package engine contains the engine API
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/ooni/probe-engine/collector"
	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/hhfm"
	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/experiment/psiphon"
	"github.com/ooni/probe-engine/experiment/sniblocking"
	"github.com/ooni/probe-engine/experiment/telegram"
	"github.com/ooni/probe-engine/experiment/tor"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/experiment/web_connectivity"
	"github.com/ooni/probe-engine/experiment/whatsapp"
	"github.com/ooni/probe-engine/internal/platform"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/version"
)

const dateFormat = "2006-01-02 15:04:05"

func formatTimeNowUTC() string {
	return time.Now().UTC().Format(dateFormat)
}

// ExperimentBuilder is an experiment builder.
type ExperimentBuilder struct {
	build         func(interface{}) *Experiment
	callbacks     model.ExperimentCallbacks
	config        interface{}
	interruptible bool
	needsInput    bool
}

// Interruptible tells you whether this is an interruptible experiment. This kind
// of experiments (e.g. ndt7) may be interrupted mid way.
func (b *ExperimentBuilder) Interruptible() bool {
	return b.interruptible
}

// NeedsInput returns whether the experiment needs input
func (b *ExperimentBuilder) NeedsInput() bool {
	return b.needsInput
}

// OptionInfo contains info about an option
type OptionInfo struct {
	Doc  string
	Type string
}

// Options returns info about all options
func (b *ExperimentBuilder) Options() (map[string]OptionInfo, error) {
	result := make(map[string]OptionInfo)
	ptrinfo := reflect.ValueOf(b.config)
	if ptrinfo.Kind() != reflect.Ptr {
		return nil, errors.New("config is not a pointer")
	}
	structinfo := ptrinfo.Elem().Type()
	if structinfo.Kind() != reflect.Struct {
		return nil, errors.New("config is not a struct")
	}
	for i := 0; i < structinfo.NumField(); i++ {
		field := structinfo.Field(i)
		result[field.Name] = OptionInfo{
			Doc:  field.Tag.Get("ooni"),
			Type: field.Type.String(),
		}
	}
	return result, nil
}

// SetOptionBool sets a bool option
func (b *ExperimentBuilder) SetOptionBool(key string, value bool) error {
	field, err := fieldbyname(b.config, key)
	if err != nil {
		return err
	}
	if field.Kind() != reflect.Bool {
		return errors.New("field is not a bool")
	}
	field.SetBool(value)
	return nil
}

// SetOptionInt sets an int option
func (b *ExperimentBuilder) SetOptionInt(key string, value int64) error {
	field, err := fieldbyname(b.config, key)
	if err != nil {
		return err
	}
	if field.Kind() != reflect.Int64 {
		return errors.New("field is not an int64")
	}
	field.SetInt(value)
	return nil
}

// SetOptionString sets a string option
func (b *ExperimentBuilder) SetOptionString(key, value string) error {
	field, err := fieldbyname(b.config, key)
	if err != nil {
		return err
	}
	if field.Kind() != reflect.String {
		return errors.New("field is not a string")
	}
	field.SetString(value)
	return nil
}

// SetCallbacks sets the interactive callbacks
func (b *ExperimentBuilder) SetCallbacks(callbacks model.ExperimentCallbacks) {
	b.callbacks = callbacks
}

func fieldbyname(v interface{}, key string) (reflect.Value, error) {
	// See https://stackoverflow.com/a/6396678/4354461
	ptrinfo := reflect.ValueOf(v)
	if ptrinfo.Kind() != reflect.Ptr {
		return reflect.Value{}, errors.New("value is not a pointer")
	}
	structinfo := ptrinfo.Elem()
	if structinfo.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("value is not a pointer to struct")
	}
	field := structinfo.FieldByName(key)
	if !field.IsValid() || !field.CanSet() {
		return reflect.Value{}, errors.New("no such field")
	}
	return field, nil
}

// NewExperiment creates the experiment
func (b *ExperimentBuilder) NewExperiment() *Experiment {
	experiment := b.build(b.config)
	experiment.callbacks = b.callbacks
	return experiment
}

// canonicalizeExperimentName allows code to provide experiment names
// in a more flexible way, where we have aliases.
func canonicalizeExperimentName(name string) string {
	switch name = strcase.ToSnake(name); name {
	case "ndt_7":
		name = "ndt" // since 2020-03-18, we use ndt7 to implement ndt by default
	default:
	}
	return name
}

func newExperimentBuilder(session *Session, name string) (*ExperimentBuilder, error) {
	factory, _ := experimentsByName[canonicalizeExperimentName(name)]
	if factory == nil {
		return nil, fmt.Errorf("no such experiment: %s", name)
	}
	builder := factory(session)
	builder.callbacks = handler.NewPrinterCallbacks(session.Logger())
	return builder, nil
}

// Experiment is an experiment instance.
type Experiment struct {
	byteCounter   *bytecounter.Counter
	callbacks     model.ExperimentCallbacks
	measurer      model.ExperimentMeasurer
	report        *collector.Report
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
		callbacks:     handler.NewPrinterCallbacks(sess.Logger()),
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

// OpenReport is an idempotent method to open a report. We assume that
// you have configured the available collectors, either manually or
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
	err = e.session.maybeLookupLocation(ctx) // this already tracks session bytes
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
	scrubErr := e.session.privacySettings.Apply(
		measurement, e.session.ProbeIP(),
	)
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

func (cb *sessionExperimentCallbacks) OnDataUsage(dloadKiB, uploadKiB float64) {
	cb.sess.byteCounter.CountKibiBytesReceived(dloadKiB)
	cb.exp.byteCounter.CountKibiBytesReceived(dloadKiB)
	cb.sess.byteCounter.CountKibiBytesSent(uploadKiB)
	cb.exp.byteCounter.CountKibiBytesSent(uploadKiB)
	cb.inner.OnDataUsage(dloadKiB, uploadKiB)
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
		DataFormatVersion:         collector.DefaultDataFormatVersion,
		Input:                     model.MeasurementTarget(input),
		MeasurementStartTime:      utctimenow.Format(dateFormat),
		MeasurementStartTimeSaved: utctimenow,
		ProbeIP:                   e.session.ProbeIP(),
		ProbeASN:                  e.session.ProbeASNString(),
		ProbeCC:                   e.session.ProbeCC(),
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
	m.AddAnnotation("engine_name", "miniooni")
	m.AddAnnotation("engine_version", version.Version)
	m.AddAnnotation("platform", platform.Name())
	return &m
}

func (e *Experiment) openReport(ctx context.Context) (err error) {
	if e.report != nil {
		return // already open
	}
	// use custom client to have proper byte accounting
	httpClient := &http.Client{
		Transport: &httptransport.ByteCountingTransport{
			RoundTripper: e.session.httpDefaultTransport, // proxy is OK
			Counter:      e.byteCounter,
		},
	}
	for _, c := range e.session.availableCollectors {
		if c.Type != "https" {
			e.session.logger.Debugf(
				"experiment: unsupported collector type: %s", c.Type,
			)
			continue
		}
		client := &collector.Client{
			BaseURL:    c.Address,
			HTTPClient: httpClient,
			Logger:     e.session.logger,
			UserAgent:  e.session.UserAgent(),
		}
		template := collector.ReportTemplate{
			DataFormatVersion: collector.DefaultDataFormatVersion,
			Format:            collector.DefaultFormat,
			ProbeASN:          e.session.ProbeASNString(),
			ProbeCC:           e.session.ProbeCC(),
			SoftwareName:      e.session.SoftwareName(),
			SoftwareVersion:   e.session.SoftwareVersion(),
			TestName:          e.testName,
			TestVersion:       e.testVersion,
		}
		e.report, err = client.OpenReport(ctx, template)
		if err == nil {
			return
		}
		e.session.logger.Debugf("experiment: collector error: %s", err.Error())
	}
	err = errors.New("All collectors failed")
	return
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

var experimentsByName = map[string]func(*Session) *ExperimentBuilder{
	"dash": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, dash.NewExperimentMeasurer(
					*config.(*dash.Config),
				))
			},
			config:        &dash.Config{},
			interruptible: true,
			needsInput:    false,
		}
	},

	"example": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example",
				))
			},
			config: &example.Config{
				Message:   "Good day from the example experiment!",
				SleepTime: int64(5 * time.Second),
			},
			interruptible: true,
			needsInput:    false,
		}
	},

	"example_with_input": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example_with_input",
				))
			},
			config: &example.Config{
				Message:   "Good day from the example with input experiment!",
				SleepTime: int64(5 * time.Second),
			},
			interruptible: true,
			needsInput:    true,
		}
	},

	// TODO(bassosimone): when we can set experiment options using the JSON
	// we need to get rid of all these multiple experiments.
	//
	// See https://github.com/ooni/probe-engine/issues/413
	"example_with_input_non_interruptible": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example_with_input_non_interruptible",
				))
			},
			config: &example.Config{
				Message:   "Good day from the example with input experiment!",
				SleepTime: int64(5 * time.Second),
			},
			interruptible: false,
			needsInput:    true,
		}
	},

	"example_with_failure": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, example.NewExperimentMeasurer(
					*config.(*example.Config), "example_with_failure",
				))
			},
			config: &example.Config{
				Message:     "Good day from the example with failure experiment!",
				ReturnError: true,
				SleepTime:   int64(5 * time.Second),
			},
			interruptible: true,
			needsInput:    false,
		}
	},

	"facebook_messenger": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, fbmessenger.NewExperimentMeasurer(
					*config.(*fbmessenger.Config),
				))
			},
			config:     &fbmessenger.Config{},
			needsInput: false,
		}
	},

	"http_header_field_manipulation": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, hhfm.NewExperimentMeasurer(
					*config.(*hhfm.Config),
				))
			},
			config:     &hhfm.Config{},
			needsInput: false,
		}
	},

	"http_invalid_request_line": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, hirl.NewExperimentMeasurer(
					*config.(*hirl.Config),
				))
			},
			config:     &hirl.Config{},
			needsInput: false,
		}
	},

	"ndt": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, ndt7.NewExperimentMeasurer(
					*config.(*ndt7.Config),
				))
			},
			config:        &ndt7.Config{},
			interruptible: true,
			needsInput:    false,
		}
	},

	"psiphon": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, psiphon.NewExperimentMeasurer(
					*config.(*psiphon.Config),
				))
			},
			config: &psiphon.Config{
				WorkDir: session.TempDir(),
			},
			needsInput: false,
		}
	},

	"sni_blocking": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, sniblocking.NewExperimentMeasurer(
					*config.(*sniblocking.Config),
				))
			},
			config: &sniblocking.Config{
				ControlSNI: "example.com",
			},
			needsInput: true,
		}
	},

	"telegram": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, telegram.NewExperimentMeasurer(
					*config.(*telegram.Config),
				))
			},
			config:     &telegram.Config{},
			needsInput: false,
		}
	},

	"tor": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, tor.NewExperimentMeasurer(
					*config.(*tor.Config),
				))
			},
			config:     &tor.Config{},
			needsInput: false,
		}
	},

	"urlgetter": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, urlgetter.NewExperimentMeasurer(
					*config.(*urlgetter.Config),
				))
			},
			config: &urlgetter.Config{
				ResolverURL: "system:///",
			},
			needsInput: true,
		}
	},

	"web_connectivity": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, web_connectivity.NewExperimentMeasurer(
					*config.(*web_connectivity.Config),
				))
			},
			config:     &web_connectivity.Config{},
			needsInput: true,
		}
	},

	"whatsapp": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *Experiment {
				return NewExperiment(session, whatsapp.NewExperimentMeasurer(
					*config.(*whatsapp.Config),
				))
			},
			config:     &whatsapp.Config{},
			needsInput: false,
		}
	},
}

// AllExperiments returns the name of all experiments
func AllExperiments() []string {
	var names []string
	for key := range experimentsByName {
		names = append(names, key)
	}
	return names
}
