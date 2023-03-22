package registry

//
// Registers the `fbmessenger' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/fbmessenger"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["facebook_messenger"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return fbmessenger.NewExperimentMeasurer(
				*config.(*fbmessenger.Config),
			)
		},
		config:      &fbmessenger.Config{},
		inputPolicy: model.InputNone,
	}
}
