package registry

//
// Registers the `simplequicping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/simplequicping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "simplequicping"
	register(canonicalName, func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return simplequicping.NewExperimentMeasurer()
			},
			canonicalName:    canonicalName,
			config:           &simplequicping.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
			newLoader:        simplequicping.NewLoader,
		}
	})
}
