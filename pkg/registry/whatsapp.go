package registry

//
// Registers the `whatsapp' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/whatsapp"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "whatsapp"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return whatsapp.NewExperimentMeasurer(
					*config.(*whatsapp.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &whatsapp.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		}
	}
}
