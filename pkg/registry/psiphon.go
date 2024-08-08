package registry

//
// Registers the `psiphon' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/psiphon"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "psiphon"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return psiphon.NewExperimentMeasurer(
					*config.(*psiphon.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &psiphon.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputOptional,
		}
	}
}
