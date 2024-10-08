package registry

//
// Registers the `hhfm' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/hhfm"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "http_header_field_manipulation"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return hhfm.NewExperimentMeasurer(
					*config.(*hhfm.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &hhfm.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		}
	}
}
