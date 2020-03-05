// Package hhfm contains the HTTP Header Field Manipulation network experiment.
package hhfm

import (
	"context"

	"github.com/ooni/probe-engine/experiment/mkhelper"
	"github.com/ooni/probe-engine/experiment/mkrunner"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

const (
	testName    = "http_header_field_manipulation"
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
		"HttpHeaderFieldManipulation", sess.SoftwareName(), sess.SoftwareVersion(),
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(), m.config.LogLevel,
	)
	err := mkhelper.Set(sess, "http-return-json-headers", "legacy", &settings)
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
