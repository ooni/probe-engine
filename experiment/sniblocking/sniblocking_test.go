package sniblocking

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
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

	experiment := NewExperiment(sess, Config{})
	if err := experiment.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer experiment.CloseReport(ctx)

	measurement, err := experiment.Measure(ctx, "www.youtube.com:443")
	if err != nil {
		t.Fatal(err)
	}

	tk := measurement.TestKeys.(*TestKeys)
	if tk.FailureWithProperSNI != nil {
		t.Fatal("expected no failure here")
	}
	if tk.FailureWithRandomSNI == nil {
		t.Fatal("expected a failure here")
	}

	if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
}

func TestFailureProperSNI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately fail it
	sess := session.New(
		log.Log, softwareName, softwareVersion, "../../testdata", nil, nil,
	)
	var measurement model.Measurement
	var config Config
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measure(ctx, sess, &measurement, callbacks, config)
	if err != nil {
		t.Fatal("expected no error here")
	}
	tk := measurement.TestKeys.(*TestKeys)
	if tk.FailureWithProperSNI == nil {
		t.Fatal("expected a failure here")
	}
	if tk.FailureWithRandomSNI == nil {
		t.Fatal("expected a failure here")
	}
}
