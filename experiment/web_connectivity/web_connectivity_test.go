package web_connectivity_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/web_connectivity"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func measureURL(
	ctx context.Context, experiment *experiment.Experiment, URL string,
) error {
	measurement, err := experiment.Measure(ctx, URL)
	if err != nil {
		return err
	}
	return experiment.SubmitMeasurement(ctx, &measurement)
}

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

	experiment := web_connectivity.NewExperiment(sess, web_connectivity.Config{})
	if err := experiment.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer experiment.CloseReport(ctx)

	for _, URL := range []string{
		"http://www.example.com/robots.txt", "http://www.google.com/humans.txt",
	} {
		if err := measureURL(ctx, experiment, URL); err != nil {
			t.Fatal(err)
		}
	}
}
