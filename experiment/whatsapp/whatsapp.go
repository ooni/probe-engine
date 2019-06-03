// Package whatsapp contains the WhatsApp network experiment.
package whatsapp

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
	testName    = "whatsapp"
	testVersion = "0.6.1"
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
		"Whatsapp", sess.SoftwareName, sess.SoftwareVersion,
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(), config.LogLevel,
	)
	settings.Options.GeoIPASNPath = sess.ASNDatabasePath()
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
