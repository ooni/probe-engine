// Package mktesting contains common code to run MK based tests.
package mktesting

import (
	"github.com/apex/log"
	engine "github.com/ooni/probe-engine"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

// Run runs the specified experiment.
func Run(input string, factory func() model.ExperimentMeasurer) error {
	if !measurementkit.Available() {
		return nil
	}
	log.SetLevel(log.DebugLevel)

	config := engine.SessionConfig{
		AssetsDir:       "../../testdata",
		Logger:          log.Log,
		SoftwareName:    softwareName,
		SoftwareVersion: softwareVersion,
		TempDir:         "../../testdata",
	}
	sess, err := engine.NewSession(config)
	if err != nil {
		return err
	}
	if err := sess.MaybeLookupBackends(); err != nil {
		return err
	}
	if err := sess.MaybeLookupLocation(); err != nil {
		return err
	}

	measurer := factory()
	experiment := engine.NewExperiment(sess, measurer)
	if err := experiment.OpenReport(); err != nil {
		return err
	}
	defer experiment.CloseReport()

	measurement, err := experiment.Measure(input)
	if err != nil {
		return err
	}
	if err := experiment.SubmitAndUpdateMeasurement(measurement); err != nil {
		return err
	}
	return nil
}
