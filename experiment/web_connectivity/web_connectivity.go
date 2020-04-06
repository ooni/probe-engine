// Package web_connectivity contains the Web Connectivity network experiment.
package web_connectivity

import (
	"context"
	"errors"

	"github.com/ooni/probe-engine/experiment/mkhelper"
	"github.com/ooni/probe-engine/experiment/mkrunner"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

const (
	testName    = "web_connectivity"
	testVersion = "0.0.1"
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
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	settings := measurementkit.NewSettings(
		"WebConnectivity", sess.SoftwareName(), sess.SoftwareVersion(),
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(), m.config.LogLevel,
	)
	settings.Options.GeoIPASNPath = sess.ASNDatabasePath()
	if measurement.Input == "" {
		return errors.New("web_connectivity: passed an empty input")
	}
	settings.Inputs = []string{string(measurement.Input)}
	err := mkhelper.Set(sess, "web-connectivity", "https", &settings)
	if err != nil {
		return err
	}
	return mkrunner.Do(
		settings, sess, measurement, callbacks, measurementkit.StartEx,
	)
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config}
}
