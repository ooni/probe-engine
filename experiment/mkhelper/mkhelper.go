// Package mkhelper contains common code to get the proper
// helper and configure it into settings.
package mkhelper

import (
	"fmt"

	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

// Set copies a specific helper from the session to MK settings.
func Set(
	sess model.ExperimentSession, name, kind string,
	settings *measurementkit.Settings,
) error {
	ths, ok := sess.GetTestHelpersByName(name)
	if !ok {
		return fmt.Errorf("No available %s test helper", name)
	}
	address := ""
	for _, th := range ths {
		if th.Type == kind {
			address = th.Address
			break
		}
	}
	if address == "" {
		return fmt.Errorf("No suitable %s test helper", name)
	}
	settings.Options.Backend = address
	return nil
}
