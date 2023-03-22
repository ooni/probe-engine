package registry

//
// Registers the `vanilla_tor' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/vanillator"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["vanilla_tor"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return vanillator.NewExperimentMeasurer(
				*config.(*vanillator.Config),
			)
		},
		config:      &vanillator.Config{},
		inputPolicy: model.InputNone,
	}
}
