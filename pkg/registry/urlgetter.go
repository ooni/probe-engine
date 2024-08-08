package registry

//
// Registers the `urlgetter' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/urlgetter"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "urlgetter"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return urlgetter.NewExperimentMeasurer(
					*config.(*urlgetter.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &urlgetter.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		}
	}
}
