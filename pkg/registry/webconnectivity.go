package registry

//
// Registers the `web_connectivity' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/webconnectivity"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["web_connectivity"] = &Factory{
		build: func(config any) model.ExperimentMeasurer {
			return webconnectivity.NewExperimentMeasurer(
				config.(webconnectivity.Config),
			)
		},
		config:        webconnectivity.Config{},
		interruptible: false,
		inputPolicy:   model.InputOrQueryBackend,
	}
}
