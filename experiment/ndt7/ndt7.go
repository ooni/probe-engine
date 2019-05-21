// Package ndt7 contains the ndt7 network experiment.
package ndt7

import (
	"context"

	upstream "github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "ndt7"
	testVersion = "0.1.0"
)

// NewReporter creates a new experiment reporter.
func NewReporter(cs *session.Session) *experiment.Reporter {
	return experiment.NewReporter(cs, testName, testVersion)
}

// TestKeys contains the test keys
type TestKeys struct {
	// Failure is the failure string
	Failure string `json:"failure"`

	// Download contains download results
	Download []spec.Measurement `json:"download"`

	// Upload contains upload results
	Upload []spec.Measurement `json:"upload"`
}

// Event is an event emitted by ndt7
type Event = spec.Measurement

// Run runs a ndt7 test
func Run(
	ctx context.Context,
	measurement *model.Measurement,
	userAgent string,
	fn func(event Event),
) error {
	testkeys := &TestKeys{}
	measurement.TestKeys = testkeys
	client := upstream.NewClient(userAgent)
	ch, err := client.StartDownload(ctx)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	for ev := range ch {
		testkeys.Download = append(testkeys.Download, ev)
		fn(ev)
	}
	ch, err = client.StartUpload(ctx)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	for ev := range ch {
		testkeys.Upload = append(testkeys.Upload, ev)
		fn(ev)
	}
	return nil
}
