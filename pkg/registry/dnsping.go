package registry

//
// Registers the `dnsping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/dnsping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["dnsping"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return dnsping.NewExperimentMeasurer(
				*config.(*dnsping.Config),
			)
		},
		config:           &dnsping.Config{},
		enabledByDefault: true,
		inputPolicy:      model.InputOrStaticDefault,
	}
}
