package registry

//
// Registers the `simple sni' experiment from the dslx tutorial.
//

import (
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/tutorial/dslx/chapter02"
)

func init() {
	AllExperiments["simple_sni"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return chapter02.NewExperimentMeasurer(
				*config.(*chapter02.Config),
			)
		},
		config:      &chapter02.Config{},
		inputPolicy: model.InputOrQueryBackend,
	}
}
