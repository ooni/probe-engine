package registry

//
// Registers the `web_connectivity' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/webconnectivity"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "web_connectivity"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config any) model.ExperimentMeasurer {
				return webconnectivity.NewExperimentMeasurer(
					*config.(*webconnectivity.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &webconnectivity.Config{},
			enabledByDefault: true,
			interruptible:    false,
			inputPolicy:      model.InputOrQueryBackend,
		}
	}
}
