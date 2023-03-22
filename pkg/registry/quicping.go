package registry

//
// Registers the `quicping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/quicping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["quicping"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return quicping.NewExperimentMeasurer(
				*config.(*quicping.Config),
			)
		},
		config:      &quicping.Config{},
		inputPolicy: model.InputStrictlyRequired,
	}
}
