package engine

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/ooni/probe-engine/model"
)

// InputPolicy describes the experiment policy with respect to input. That is
// whether it requires input, optionally accepts input, does not want input.
type InputPolicy string

const (
	// InputRequired indicates that the experiment requires
	// external input to run. If this input is not provided to
	// the experiment, it will not know what to do.
	InputRequired = InputPolicy("required")

	// InputOptional indicates that the experiment handles input,
	// if any; otherwise it fetchs input/uses a default.
	InputOptional = InputPolicy("optional")

	// InputNone indicates that the experiment does not want any
	// input and ignores the input if provided with it.
	InputNone = InputPolicy("none")
)

// ExperimentBuilder is an experiment builder.
type ExperimentBuilder struct {
	build         func(interface{}) *Experiment
	callbacks     model.ExperimentCallbacks
	config        interface{}
	inputPolicy   InputPolicy
	interruptible bool
}

// Interruptible tells you whether this is an interruptible experiment. This kind
// of experiments (e.g. ndt7) may be interrupted mid way.
func (b *ExperimentBuilder) Interruptible() bool {
	return b.interruptible
}

// InputPolicy returns the experiment input policy
func (b *ExperimentBuilder) InputPolicy() InputPolicy {
	return b.inputPolicy
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
	builder.callbacks = model.NewPrinterCallbacks(session.Logger())
	return builder, nil
}
