package registry

//
// Registers the `tlstool' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/tlstool"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["tlstool"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return tlstool.NewExperimentMeasurer(
				*config.(*tlstool.Config),
			)
		},
		config:           &tlstool.Config{},
		enabledByDefault: true,
		inputPolicy:      model.InputOrQueryBackend,
	}
}
