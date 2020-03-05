// Package whatsapp contains the WhatsApp network experiment.
package whatsapp

import (
	"context"

	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/mkrunner"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/model2"
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

type measurer struct {
	config Config
}

func (m *measurer) ExperimentName() string {
	return testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) Run(
	ctx context.Context, sess *session.Session,
	measurement *model.Measurement, callbacks handler.Callbacks,
) error {
	settings := measurementkit.NewSettings(
		"Whatsapp", sess.SoftwareName, sess.SoftwareVersion,
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(), m.config.LogLevel,
	)
	settings.Options.GeoIPASNPath = sess.ASNDatabasePath()
	return mkrunner.Do(
		settings, sess, measurement, callbacks, measurementkit.StartEx,
	)
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model2.ExperimentMeasurer {
	return &measurer{config: config}
}
