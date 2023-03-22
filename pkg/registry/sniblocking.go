package registry

//
// Registers the `sniblocking' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/sniblocking"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["sni_blocking"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return sniblocking.NewExperimentMeasurer(
				*config.(*sniblocking.Config),
			)
		},
		config:      &sniblocking.Config{},
		inputPolicy: model.InputOrQueryBackend,
	}
}
