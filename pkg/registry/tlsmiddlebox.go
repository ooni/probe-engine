package registry

//
// Registers the `tlsmiddlebox' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/tlsmiddlebox"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "tlsmiddlebox"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return tlsmiddlebox.NewExperimentMeasurer(
					*config.(*tlsmiddlebox.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &tlsmiddlebox.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputStrictlyRequired,
		}
	}
}
