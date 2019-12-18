package example_test

import (
	"context"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/example"
	"github.com/ooni/probe-engine/internal/kvstore"
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
		"../../testdata", kvstore.NewMemoryKeyValueStore(),
	)
	if err := sess.MaybeLookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	experiment := example.NewExperiment(
		sess, example.Config{SleepTime: int64(2 * time.Second)},
	)
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

func TestIntegrationFailure(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	sess := session.New(
		log.Log, softwareName, softwareVersion, "../../testdata", nil, nil,
		"../../testdata", kvstore.NewMemoryKeyValueStore(),
	)
	experiment := example.NewExperiment(
		sess, example.Config{
			SleepTime:   int64(2 * time.Second),
			ReturnError: true,
		},
	)
	if _, err := experiment.Measure(ctx, ""); err == nil {
		t.Fatal("expected an error here")
	}
}
