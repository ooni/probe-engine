package registry

//
// Registers the `signal' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/signal"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "signal"
	register(canonicalName, func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return signal.NewExperimentMeasurer()
			},
			canonicalName:    canonicalName,
			config:           &signal.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
			newLoader:        signal.NewLoader,
		}
	})
}
