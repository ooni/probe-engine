package registry

//
// Registers the `run' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/run"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["run"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return run.NewExperimentMeasurer(
				*config.(*run.Config),
			)
		},
		config:           &run.Config{},
		enabledByDefault: true,
		inputPolicy:      model.InputStrictlyRequired,
	}
}
