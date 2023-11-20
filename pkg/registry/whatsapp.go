package registry

//
// Registers the `whatsapp' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/whatsapp"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["whatsapp"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return whatsapp.NewExperimentMeasurer(
				*config.(*whatsapp.Config),
			)
		},
		config:           &whatsapp.Config{},
		enabledByDefault: true,
		inputPolicy:      model.InputNone,
	}
}
