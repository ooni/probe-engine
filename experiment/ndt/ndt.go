// Package ndt contains the ndt network experiment.
package ndt

import (
	"context"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "ndt"
	testVersion = "0.1.0"
)

// NewReporter creates a new experiment reporter.
func NewReporter(cs *session.Session) *experiment.Reporter {
	return experiment.NewReporter(cs, testName, testVersion)
}

// Run runs a ndt test
func Run(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
) error {
	settings := measurementkit.NewSettings(
		"Ndt", sess.SoftwareName, sess.SoftwareVersion,
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(),
	)
	out, err := measurementkit.StartEx(settings, sess.Logger)
	if err != nil {
		return err
	}
	for range out {
		// Drain
	}
	return nil
}
