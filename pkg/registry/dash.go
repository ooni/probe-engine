package registry

//
// Registers the `dash' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/dash"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "dash"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return dash.NewExperimentMeasurer(
					*config.(*dash.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &dash.Config{},
			enabledByDefault: true,
			interruptible:    true,
			inputPolicy:      model.InputNone,
		}
	}
}
