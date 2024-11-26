package registry

//
// Registers the `hirl' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/hirl"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "http_invalid_request_line"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return hirl.NewExperimentMeasurer(
					*config.(*hirl.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &hirl.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		}
	}
}
