package registry

//
// Registers the `sniblocking' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/sniblocking"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "sni_blocking"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return sniblocking.NewExperimentMeasurer(
					*config.(*sniblocking.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &sniblocking.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputOrQueryBackend,
		}
	}
}
