package registry

//
// Registers the `simplequicping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/simplequicping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["simplequicping"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return simplequicping.NewExperimentMeasurer(
				*config.(*simplequicping.Config),
			)
		},
		config:           &simplequicping.Config{},
		enabledByDefault: true,
		inputPolicy:      model.InputStrictlyRequired,
	}
}
