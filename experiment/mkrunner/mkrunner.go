// Package mkrunner contains code to run an MK based test
package mkrunner

import (
	"errors"

	"github.com/ooni/probe-engine/experiment/mkevent"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

// Do runs a specific measurement-kit based experiment. You should
// pass measurementkit.StartEx as startEx in the common case.
func Do(
	settings measurementkit.Settings,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
	startEx func(
		settings measurementkit.Settings, logger model.Logger,
	) (<-chan measurementkit.Event, error),
) error {
	out, err := startEx(settings, sess.Logger())
	if err != nil {
		return err
	}
	// Try to increase reliability a bit. MK does not do a good job at
	// that, traditionally. Declare an experiment run failed if we didn't
	// at least see a `measurement` event.
	err = errors.New("did not see any measurement event")
	for ev := range out {
		if ev.Key == "measurement" {
			err = nil
		}
		mkevent.Handle(sess, measurement, ev, callbacks)
	}
	return err
}

// DoNothingStartEx is a replacement for measurementkit.StartEx that
// does nearly nothing apart emitting a fake measurement.
func DoNothingStartEx(
	settings measurementkit.Settings, logger model.Logger,
) (<-chan measurementkit.Event, error) {
	out := make(chan measurementkit.Event)
	go func() {
		defer close(out)
		out <- measurementkit.Event{
			Key: "measurement",
			Value: measurementkit.EventValue{
				JSONStr: "{}",
			},
		}
	}()
	return out, nil
}

// FailingStartEx is a replacement for measurementkit.StartEx that
// returns an error immediately.
func FailingStartEx(
	settings measurementkit.Settings, logger model.Logger,
) (<-chan measurementkit.Event, error) {
	return nil, errors.New("fail immediately")
}
