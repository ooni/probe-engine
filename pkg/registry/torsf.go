package registry

//
// Registers the `torsf' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/torsf"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "torsf"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return torsf.NewExperimentMeasurer(
					*config.(*torsf.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &torsf.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputNone,
		}
	}
}
