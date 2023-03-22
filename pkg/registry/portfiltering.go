package registry

//
// Registers the 'portfiltering' experiment
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/portfiltering"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["portfiltering"] = &Factory{
		build: func(config any) model.ExperimentMeasurer {
			return portfiltering.NewExperimentMeasurer(
				config.(portfiltering.Config),
			)
		},
		config:        portfiltering.Config{},
		interruptible: false,
		inputPolicy:   model.InputNone,
	}
}
