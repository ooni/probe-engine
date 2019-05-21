// Package web_connectivity contains the Web Connectivity network experiment.
package web_connectivity

import (
	"context"
	"errors"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/mkevent"
	"github.com/ooni/probe-engine/experiment/mkhelper"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "web_connectivity"
	testVersion = "0.0.1"
)

// Config contains the experiment config.
type Config struct{}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
) error {
	settings := measurementkit.NewSettings(
		"WebConnectivity", sess.SoftwareName, sess.SoftwareVersion,
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(),
	)
	settings.Options.GeoIPASNPath = sess.ASNDatabasePath()
	if measurement.Input == "" {
		return errors.New("web_connectivity: passed an empty input")
	}
	settings.Inputs = []string{measurement.Input}
	err := mkhelper.Set(sess, "web-connectivity", "https", &settings)
	if err != nil {
		return err
	}
	out, err := measurementkit.StartEx(settings, sess.Logger)
	if err != nil {
		return err
	}
	for ev := range out {
		mkevent.Handle(sess, measurement, ev)
	}
	return nil
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	return experiment.New(sess, testName, testVersion, measure)
}
