// Package webconnectivity implements OONI's Web Connectivity experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-017-web-connectivity.md
package webconnectivity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
)

const (
	testName    = "web_connectivity"
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains webconnectivity test keys.
type TestKeys struct {
	// measurement
	urlgetter.TestKeys

	// contextual information
	ClientResolver string `json:"client_resolver"`

	// control
	ControlFailure *string         `json:"control_failure"`
	ControlRequest ControlRequest  `json:"x_control_request"` // not in the spec
	Control        ControlResponse `json:"control"`
}

// Measurer performs the measurement.
type Measurer struct {
	Config Config
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return Measurer{Config: config}
}

// ExperimentName implements ExperimentMeasurer.ExperExperimentName.
func (m Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion implements ExperimentMeasurer.ExperExperimentVersion.
func (m Measurer) ExperimentVersion() string {
	return testVersion
}

var (
	// ErrNoAvailableTestHelpers is emitted when there are no available test helpers.
	ErrNoAvailableTestHelpers = errors.New("no available helpers")

	// ErrNoInput indicates that no input was provided
	ErrNoInput = errors.New("no input provided")

	// ErrUnsupportedInput indicates that the input URL scheme is unsupported.
	ErrUnsupportedInput = errors.New("unsupported input scheme")
)

// Run implements ExperimentMeasurer.Run.
func (m Measurer) Run(
	ctx context.Context,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	urlgetter.RegisterExtensions(measurement)
	tk := new(TestKeys)
	measurement.TestKeys = tk
	tk.ClientResolver = sess.ResolverIP()
	if measurement.Input == "" {
		return ErrNoInput
	}
	if strings.HasPrefix(string(measurement.Input), "http://") != true &&
		strings.HasPrefix(string(measurement.Input), "https://") != true {
		return ErrUnsupportedInput
	}
	// 1. find test helper
	testhelpers, _ := sess.GetTestHelpersByName("web-connectivity")
	var testhelper *model.Service
	for _, th := range testhelpers {
		if th.Type == "https" {
			testhelper = &th
			break
		}
	}
	if testhelper == nil {
		return ErrNoAvailableTestHelpers
	}
	measurement.TestHelpers = map[string]interface{}{
		"backend": testhelper,
	}
	// 2. perform the measurement
	tk.TestKeys = Measure(ctx, sess, measurement.Input)
	// 3. contact the control
	tk.ControlRequest = NewControlRequest(measurement.Input, tk.TestKeys)
	var err error
	tk.Control, err = Control(ctx, sess, testhelper.Address, tk.ControlRequest)
	tk.ControlFailure = archival.NewFailure(err)
	// 4. compare measurement to control
	Analyze(tk)
	return nil
}
