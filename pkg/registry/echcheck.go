package registry

//
// Registers the `echcheck' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/echcheck"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["echcheck"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return echcheck.NewExperimentMeasurer(
				*config.(*echcheck.Config),
			)
		},
		config:      &echcheck.Config{},
		inputPolicy: model.InputOptional,
	}
}
