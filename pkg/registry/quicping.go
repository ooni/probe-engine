package registry

//
// Registers the `quicping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/quicping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "quicping"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return quicping.NewExperimentMeasurer(
					*config.(*quicping.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &quicping.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		}
	}
}
