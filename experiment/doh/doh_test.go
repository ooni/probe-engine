package doh_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/doh"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()

	sess := session.New(
		log.Log, softwareName, softwareVersion, "../../testdata", nil, nil,
	)
	if err := sess.MaybeLookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	experiment := doh.NewExperiment(sess, doh.Config{})
	if err := experiment.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer experiment.CloseReport(ctx)

	measurement, err := experiment.Measure(ctx, "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
}
