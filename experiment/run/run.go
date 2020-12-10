// Package run contains code to run other experiments.
package run

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ooni/probe-engine/experiment/dnscheck"
	"github.com/ooni/probe-engine/model"
)

// Config contains settings.
type Config struct{}

// Measurer runs the measurement.
type Measurer struct{}

// ExperimentName implements ExperimentMeasurer.ExperimentName.
func (Measurer) ExperimentName() string {
	return "run"
}

// ExperimentVersion implements ExperimentMeasurer.ExperimentVersion.
func (Measurer) ExperimentVersion() string {
	return "0.1.0"
}

// StructuredInput contains structured input for this experiment.
type StructuredInput struct {
	// DNSCheck contains settings for the dnscheck experiment.
	DNSCheck dnscheck.Config `json:"dns_check"`

	// Name is the name of the experiment to run.
	Name string `json:"name"`

	// Input is the input for this experiment.
	Input string `json:"input"`
}

// Run implements ExperimentMeasurer.ExperimentVersion.
func (Measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	var input StructuredInput
	if err := json.Unmarshal([]byte(measurement.Input), &input); err != nil {
		return err
	}
	mainfunc, found := table[input.Name]
	if !found {
		return fmt.Errorf("no such experiment: %s", input.Name)
	}
	return mainfunc(ctx, input, sess, measurement, callbacks)
}

// GetSummaryKeys implements ExperimentMeasurer.GetSummaryKeys
func (Measurer) GetSummaryKeys(*model.Measurement) (interface{}, error) {
	// TODO(bassosimone): we could extend this interface to call the
	// specific GetSummaryKeys of the experiment we're running.
	return dnscheck.SummaryKeys{IsAnomaly: false}, nil
}

// NewExperimentMeasurer creates a new model.ExperimentMeasurer
// implementing the run experiment.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return Measurer{}
}
