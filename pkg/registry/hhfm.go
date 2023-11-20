package registry

//
// Registers the `hhfm' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/hhfm"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["http_header_field_manipulation"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return hhfm.NewExperimentMeasurer(
				*config.(*hhfm.Config),
			)
		},
		config:           &hhfm.Config{},
		enabledByDefault: true,
		inputPolicy:      model.InputNone,
	}
}
