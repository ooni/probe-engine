package fbmessenger_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()

	sess := session.New(log.Log, softwareName, softwareVersion, "../../testdata")
	if err := sess.LookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	experiment := fbmessenger.NewExperiment(sess, fbmessenger.Config{})
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
