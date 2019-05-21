package ndt7_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()

	sess := session.New(log.Log, softwareName, softwareVersion)
	sess.WorkDir = "../../testdata"
	if err := sess.LookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	reporter := ndt7.NewReporter(sess)
	if err := reporter.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer reporter.CloseReport(ctx)

	measurement := reporter.NewMeasurement("")
	if err := ndt7.Run(
		ctx, &measurement, sess.UserAgent(), func(event ndt7.Event) {
			data, err := json.Marshal(event)
			if err != nil {
				panic(err) // should not happen
			}
			log.Debug(string(data))
		}); err != nil {
		t.Fatal(err)
	}
	if err := reporter.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
}
