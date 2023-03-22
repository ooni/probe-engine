package registry

//
// Registers the `tcpping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/tcpping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["tcpping"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return tcpping.NewExperimentMeasurer(
				*config.(*tcpping.Config),
			)
		},
		config:      &tcpping.Config{},
		inputPolicy: model.InputStrictlyRequired,
	}
}
