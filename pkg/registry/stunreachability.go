package registry

//
// Registers the `stunreachability' experiment.
//

import (
	"github.com/ooni/probe-engine/pkg/experiment/stunreachability"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	const canonicalName = "stunreachability"
	register(canonicalName, func() *Factory {
		return &Factory{
			build: func(config interface{}) model.ExperimentMeasurer {
				return stunreachability.NewExperimentMeasurer()
			},
			canonicalName:    canonicalName,
			config:           &stunreachability.Config{},
			enabledByDefault: true,
			inputPolicy:      model.InputOrStaticDefault,
			newLoader:        stunreachability.NewLoader,
		}
	})
}
