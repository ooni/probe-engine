// Package example contains a simple example of experiment.
package example

import (
	"context"
	"errors"
	"time"

	"github.com/ooni/probe-engine/model"
)

const testVersion = "0.0.1"

// Config contains the experiment config.
//
// This contains all the settings that user can set to modify the behaviour
// of this experiment.
type Config struct {
	Message     string `ooni:"Message to emit at test completion"`
	ReturnError bool   `ooni:"Toogle to return a mocked error"`
	SleepTime   int64  `ooni:"Amount of time to sleep for"`
}

// TestKeys contains the experiment's result.
//
// This is what will end up into the Measurement.TestKeys field
// when you run this experiment.
// In other words, the variables in this struct will be
// the particular results of this experiment.
type TestKeys struct {
	Success bool `json:"success"`
}

type measurer struct {
	config   Config
	testName string
}

func (m *measurer) ExperimentName() string {
	return m.testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	var err error
	if m.config.ReturnError {
		err = errors.New("mocked error")
	}
	testkeys := &TestKeys{Success: err == nil}
	measurement.TestKeys = testkeys
	ctx, cancel := context.WithTimeout(ctx, time.Duration(m.config.SleepTime))
	defer cancel()
	<-ctx.Done()
	sess.Logger().Warnf("example: remember to drink: %s", "water is key to survival")
	callbacks.OnProgress(1.0, m.config.Message)
	callbacks.OnDataUsage(0, 0)
	return err
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config, testName string) model.ExperimentMeasurer {
	return &measurer{config: config, testName: testName}
}
