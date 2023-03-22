package registry

//
// Registers the `tlsping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/tlsping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["tlsping"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return tlsping.NewExperimentMeasurer(
				*config.(*tlsping.Config),
			)
		},
		config:      &tlsping.Config{},
		inputPolicy: model.InputStrictlyRequired,
	}
}
