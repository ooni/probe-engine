// Package hirl contains the HTTP Invalid Request Line network experiment.
package hirl

import (
	"context"
	"fmt"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/mkevent"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "http_invalid_request_line"
	testVersion = "0.0.3"
)

// Config contains the experiment config.
type Config struct{}

func gethelper(sess *session.Session, name, kind string) (string, error) {
	ths, ok := sess.AvailableTestHelpers[name]
	if !ok {
		return "", fmt.Errorf("No available %s test helper", name)
	}
	address := ""
	for _, th := range ths {
		if th.Type == kind {
			address = th.Address
			break
		}
	}
	if address == "" {
		return "", fmt.Errorf("No suitable %s test helper", name)
	}
	return address, nil
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
) error {
	settings := measurementkit.NewSettings(
		"HttpInvalidRequestLine", sess.SoftwareName, sess.SoftwareVersion,
		sess.CABundlePath(), sess.ProbeASNString(), sess.ProbeCC(),
		sess.ProbeIP(), sess.ProbeNetworkName(),
	)
	helper, err := gethelper(sess, "tcp-echo", "legacy")
	if err != nil {
		return err
	}
	settings.Options.Backend = helper
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
