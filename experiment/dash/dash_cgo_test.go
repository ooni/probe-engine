// +build cgo

package dash_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	if !measurementkit.Available() {
		t.Skip("Measurement Kit not available; skipping")
	}
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()

	sess := session.New(
		log.Log, softwareName, softwareVersion, "../../testdata", nil,
		"../../testdata", kvstore.NewMemoryKeyValueStore(),
	)
	if err := sess.MaybeLookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	experiment := dash.NewExperiment(sess, dash.Config{})
	if err := experiment.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer experiment.CloseReport(ctx)

	measurement, err := experiment.Measure(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
}
