package registry

//
// Registers the `dnsping' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/dnsping"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "dnsping"
	AllExperiments[canonicalName] = func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return dnsping.NewExperimentMeasurer(
					*config.(*dnsping.Config),
				)
			},
			canonicalName:    canonicalName,
			config:           &dnsping.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputOrStaticDefault,
		}
	}
}
