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

// Config contains the experiment settings
type Config struct{}

// TestKeys contains the test keys
type TestKeys struct {
	// Failure is the failure string
	Failure string `json:"failure"`

	// Download contains download results
	Download []spec.Measurement `json:"download"`

	// Upload contains upload results
	Upload []spec.Measurement `json:"upload"`
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
) error {
	testkeys := &TestKeys{}
	measurement.TestKeys = testkeys
	client := upstream.NewClient(sess.UserAgent())
	ch, err := client.StartDownload(ctx)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	for ev := range ch {
		testkeys.Download = append(testkeys.Download, ev)
		sess.Logger.Debugf("%+v", ev)
	}
	ch, err = client.StartUpload(ctx)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	for ev := range ch {
		testkeys.Upload = append(testkeys.Upload, ev)
		sess.Logger.Debugf("%+v", ev)
	}
	return nil
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	return experiment.New(sess, testName, testVersion, measure)
}
