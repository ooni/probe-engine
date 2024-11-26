package registry

//
// Registers the `riseupvpn' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/riseupvpn"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "riseupvpn"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return riseupvpn.NewExperimentMeasurer(
					*config.(*riseupvpn.Config),
				)
			},
			canonicalName: canonicalName,
			config:        &riseupvpn.Config{},
			inputPolicy:   model.InputNone,
		}
	}
}
