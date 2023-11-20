package registry

//
// Registers the `torsf' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/torsf"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["torsf"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return torsf.NewExperimentMeasurer(
				*config.(*torsf.Config),
			)
		},
		config:           &torsf.Config{},
		enabledByDefault: false,
		inputPolicy:      model.InputNone,
	}
}
