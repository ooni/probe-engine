// Package ndt7 contains the ndt7 nettest
package ndt7

import (
	"context"

	upstream "github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/nettest"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "ndt7"
	testVersion = "0.1.0"
)

// NewNettest creates a new ndt7 nettest.
func NewNettest(cs *session.Session) *nettest.Nettest {
	return nettest.New(cs, testName, testVersion)
}

// testKeys contains the test keys
type testKeys struct {
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
	fn func(event Event),
) error {
	testkeys := &testKeys{}
	measurement.TestKeys = testkeys
	client := upstream.NewClient("ooniprobe-example/0.0.1")
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
