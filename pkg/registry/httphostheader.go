package registry

//
// Registers the `httphostheader' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/httphostheader"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["http_host_header"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return httphostheader.NewExperimentMeasurer(
				*config.(*httphostheader.Config),
			)
		},
		config:      &httphostheader.Config{},
		inputPolicy: model.InputOrQueryBackend,
	}
}
