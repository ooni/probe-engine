// Package engine contains the engine API
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/hhfm"
	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/experiment/ndt"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/experiment/psiphon"
	"github.com/ooni/probe-engine/experiment/sniblocking"
	"github.com/ooni/probe-engine/experiment/telegram"
	"github.com/ooni/probe-engine/experiment/tor"
	"github.com/ooni/probe-engine/experiment/web_connectivity"
	"github.com/ooni/probe-engine/experiment/whatsapp"
	"github.com/ooni/probe-engine/model"
)

// Callbacks contains event handling callbacks
//
// This is a copy of experiment/handler.Callbacks. Go will make sure
// the interface will match for us. This means we can have this set of
// callbacks as part of the toplevel engine API.
type Callbacks interface {
	// OnDataUsage provides information about data usage.
	OnDataUsage(dloadKiB, uploadKiB float64)

	// OnProgress provides information about an experiment progress.
	OnProgress(percentage float64, message string)
}

// ExperimentBuilder is an experiment builder.
type ExperimentBuilder struct {
	build      func(interface{}) *experiment.Experiment
	callbacks  Callbacks
	config     interface{}
	needsInput bool
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
func (b *ExperimentBuilder) SetCallbacks(callbacks Callbacks) {
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

// Build builds the experiment
func (b *ExperimentBuilder) Build() *Experiment {
	experiment := b.build(b.config)
	experiment.Callbacks = b.callbacks
	return &Experiment{
		experiment: experiment,
		name:       experiment.TestName,
	}
}

// canonicalizeExperimentName allows code to provide experiment names
// in a more flexible way. There is a special case for ndt7 where we
// get `ndt_7` but we would actually want `ndt7`.
func canonicalizeExperimentName(name string) string {
	switch name = strcase.ToSnake(name); name {
	case "ndt_7":
		name = "ndt7"
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
	builder.callbacks = handler.NewPrinterCallbacks(session.session.Logger)
	return builder, nil
}

// Experiment is an experiment instance.
type Experiment struct {
	experiment *experiment.Experiment
	name       string
}

// Name returns the experiment name.
func (e *Experiment) Name() string {
	return e.name
}

// OpenReport is an idempotent method to open a report. We assume that
// you have configured the available collectors, either manually or
// through using the session's MaybeLookupBackends method.
func (e *Experiment) OpenReport() error {
	ctx := context.Background()
	return e.experiment.OpenReport(ctx)
}

// ReportID returns the open reportID, if we have opened a report
// successfully before, or an empty string, otherwise.
func (e *Experiment) ReportID() string {
	return e.experiment.ReportID()
}

// Measure performs a measurement with input. We assume that you have
// configured the available test helpers, either manually or by calling
// the session's MaybeLookupBackends() method.
func (e *Experiment) Measure(input string) (*Measurement, error) {
	ctx := context.Background()
	measurement, err := e.experiment.Measure(ctx, input)
	// Note: the experiment returns a measurement and not a pointer
	// therefore we can always safely wrap what we've got. This is
	// in line with knowing also from the measurement what was wrong.
	return &Measurement{m: measurement}, err
}

// LoadMeasurement loads a measurement from a byte stream. The measurement
// must be a measurement for this experiment.
func (e *Experiment) LoadMeasurement(data []byte) (*Measurement, error) {
	var measurement model.Measurement
	if err := json.Unmarshal(data, &measurement); err != nil {
		return nil, err
	}
	if measurement.TestName != e.Name() {
		return nil, errors.New("not a measurement for this experiment")
	}
	return &Measurement{m: measurement}, nil
}

// SubmitAndUpdateMeasurement submits a measurement and updates the
// fields whose value has changed as part of the submission.
func (e *Experiment) SubmitAndUpdateMeasurement(measurement *Measurement) error {
	return e.experiment.SubmitMeasurement(context.Background(), &measurement.m)
}

// SaveMeasurement saves the measurement at the specified path.
func (e *Experiment) SaveMeasurement(measurement *Measurement, path string) error {
	return e.experiment.SaveMeasurement(measurement.m, path)
}

// CloseReport is an idempotent method that closes and open report
// if one has previously been opened, otherwise it does nothing.
func (e *Experiment) CloseReport() error {
	return e.experiment.CloseReport(context.Background())
}

// Measurement is a OONI measurement
type Measurement struct {
	m model.Measurement
}

// MarshalJSON marshals the measurement as JSON
func (m *Measurement) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.m)
}

// AddAnnotations adds annotation to the measurement
func (m *Measurement) AddAnnotations(annotations map[string]string) {
	m.m.AddAnnotations(annotations)
}

// MakeGenericTestKeys casts the m.TestKeys to a map[string]interface{}.
//
// Ideally, all tests should have a clear Go structure, well defined, that
// will be stored in m.TestKeys as an interface. This is not already the
// case and it's just valid for tests written in Go. Until all tests will
// be written in Go, we'll keep this glue here to make sure we convert from
// the engine format to the cli format.
//
// This function will first attempt to cast directly to map[string]interface{},
// which is possible for MK tests, and then use JSON serialization and
// de-serialization only if that's required.
func (m *Measurement) MakeGenericTestKeys() (map[string]interface{}, error) {
	return m.makeGenericTestKeys(json.Marshal)
}

func (m *Measurement) makeGenericTestKeys(
	marshal func(v interface{}) ([]byte, error),
) (map[string]interface{}, error) {
	if result, ok := m.m.TestKeys.(map[string]interface{}); ok {
		return result, nil
	}
	data, err := marshal(m.m.TestKeys)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

var experimentsByName = map[string]func(*Session) *ExperimentBuilder{
	"dash": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return dash.NewExperiment(session.session, *config.(*dash.Config))
			},
			config:     &dash.Config{},
			needsInput: false,
		}
	},

	"example": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return example.NewExperiment(session.session, *config.(*example.Config))
			},
			config: &example.Config{
				Message:   "Good day from the example experiment!",
				SleepTime: int64(2 * time.Second),
			},
			needsInput: false,
		}
	},

	"facebook_messenger": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return fbmessenger.NewExperiment(
					session.session, *config.(*fbmessenger.Config),
				)
			},
			config:     &fbmessenger.Config{},
			needsInput: false,
		}
	},

	"http_header_field_manipulation": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return hhfm.NewExperiment(session.session, *config.(*hhfm.Config))
			},
			config:     &hhfm.Config{},
			needsInput: false,
		}
	},

	"http_invalid_request_line": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return hirl.NewExperiment(session.session, *config.(*hirl.Config))
			},
			config:     &hirl.Config{},
			needsInput: false,
		}
	},

	"ndt": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return ndt.NewExperiment(session.session, *config.(*ndt.Config))
			},
			config:     &ndt.Config{},
			needsInput: false,
		}
	},

	"ndt7": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return ndt7.NewExperiment(session.session, *config.(*ndt7.Config))
			},
			config:     &ndt7.Config{},
			needsInput: false,
		}
	},

	"psiphon": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return psiphon.NewExperiment(session.session, *config.(*psiphon.Config))
			},
			config: &psiphon.Config{
				WorkDir: session.session.TempDir,
			},
			needsInput: false,
		}
	},

	"sni_blocking": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return sniblocking.NewExperiment(session.session, *config.(*sniblocking.Config))
			},
			config: &sniblocking.Config{
				ControlSNI: "ps.ooni.io",
			},
			needsInput: true,
		}
	},

	"telegram": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return telegram.NewExperiment(session.session, *config.(*telegram.Config))
			},
			config:     &telegram.Config{},
			needsInput: false,
		}
	},

	"tor": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return tor.NewExperiment(session.session, *config.(*tor.Config))
			},
			config:     &tor.Config{},
			needsInput: false,
		}
	},

	"web_connectivity": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return web_connectivity.NewExperiment(
					session.session, *config.(*web_connectivity.Config),
				)
			},
			config:     &web_connectivity.Config{},
			needsInput: true,
		}
	},

	"whatsapp": func(session *Session) *ExperimentBuilder {
		return &ExperimentBuilder{
			build: func(config interface{}) *experiment.Experiment {
				return whatsapp.NewExperiment(
					session.session, *config.(*whatsapp.Config),
				)
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
