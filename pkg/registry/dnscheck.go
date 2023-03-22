package registry

//
// Registers the `dnscheck' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/dnscheck"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["dnscheck"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return dnscheck.NewExperimentMeasurer(
				*config.(*dnscheck.Config),
			)
		},
		config:      &dnscheck.Config{},
		inputPolicy: model.InputOrStaticDefault,
	}
}
