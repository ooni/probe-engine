// Package example contains a simple example of experiment.
package example

import (
	"context"
	"errors"
	"time"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "example"
	testVersion = "0.0.1"
)

// Config contains the experiment config.
//
// This contains all the settings that user can set to modify the behaviour
// of this experiment.
type Config struct {
	ReturnError bool   `ooni:"Toogle to return a mocked error"`
	Message     string `ooni:"Message to emit at test completion"`
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

// measure is the main function of each experiment,
// that is called by the NewExperiment.
func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	var err error
	if config.ReturnError {
		err = errors.New("mocked error")
	}
	testkeys := &TestKeys{Success: err == nil}
	measurement.TestKeys = testkeys
	time.Sleep(time.Duration(config.SleepTime))
	callbacks.OnProgress(1.0, config.Message)
	callbacks.OnDataUsage(0, 0)
	return err
}

// NewExperiment creates a new experiment.
//
// This is the function that you call to create an instance of this experiment.
// Once you have created an instance, you can use directly the
// generic experiment API.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	return experiment.New(
		sess, testName, testVersion,
		func(
			ctx context.Context,
			sess *session.Session,
			measurement *model.Measurement,
			callbacks handler.Callbacks,
		) error {
			return measure(ctx, sess, measurement, callbacks, config)
		})
}
