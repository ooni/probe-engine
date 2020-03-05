// Package mktesting contains common code to run MK based tests.
package mktesting

import (
	"context"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model2"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

// Run runs the specified experiment.
func Run(input string, factory func() model2.ExperimentMeasurer) error {
	if !measurementkit.Available() {
		return nil
	}
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()

	sess := session.New(
		log.Log, softwareName, softwareVersion, "../../testdata", nil,
		"../../testdata", kvstore.NewMemoryKeyValueStore(),
	)
	if err := sess.MaybeLookupBackends(ctx); err != nil {
		return err
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		return err
	}

	measurer := factory()
	experiment := experiment.New(
		sess, measurer.ExperimentName(),
		measurer.ExperimentVersion(), measurer,
	)
	if err := experiment.OpenReport(ctx); err != nil {
		return err
	}
	defer experiment.CloseReport(ctx)

	measurement, err := experiment.Measure(ctx, input)
	if err != nil {
		return err
	}
	if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
		return err
	}
	return nil
}
