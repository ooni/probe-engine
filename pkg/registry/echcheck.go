package registry

//
// Registers the `echcheck' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/echcheck"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "echcheck"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return echcheck.NewExperimentMeasurer(
					*config.(*echcheck.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &echcheck.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputOptional,
		}
	}
}