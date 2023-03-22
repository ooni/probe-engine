package registry

//
// Registers the `stunreachability' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/stunreachability"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["stunreachability"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return stunreachability.NewExperimentMeasurer(
				*config.(*stunreachability.Config),
			)
		},
		config:      &stunreachability.Config{},
		inputPolicy: model.InputOrStaticDefault,
	}
}
