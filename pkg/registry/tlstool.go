package registry

//
// Registers the `tlstool' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/tlstool"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "tlstool"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return tlstool.NewExperimentMeasurer(
					*config.(*tlstool.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &tlstool.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
		}
	}
}
