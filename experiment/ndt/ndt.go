// Package ndt contains the ndt network experiment.
package ndt

import (
	"context"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/mkevent"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "ndt"
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct {
	// LogLevel is the MK log level. Empty implies "WARNING".
	LogLevel string
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	settings := measurementkit.NewSettings(
		"Ndt", sess.SoftwareName, sess.SoftwareVersion,
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(), config.LogLevel,
	)
	out, err := measurementkit.StartEx(settings, sess.Logger)
	if err != nil {
		return err
	}
	for ev := range out {
		mkevent.Handle(sess, measurement, ev, callbacks)
	}
	return nil
}

// NewExperiment creates a new experiment.
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
