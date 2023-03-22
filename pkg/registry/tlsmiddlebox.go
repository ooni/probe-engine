package registry

//
// Registers the `tlsmiddlebox' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/tlsmiddlebox"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["tlsmiddlebox"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return tlsmiddlebox.NewExperimentMeasurer(
				*config.(*tlsmiddlebox.Config),
			)
		},
		config:      &tlsmiddlebox.Config{},
		inputPolicy: model.InputStrictlyRequired,
	}
}
