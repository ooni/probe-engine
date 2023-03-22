package registry

//
// Registers the `example' experiment.
//

import (
	"time"

	"github.com/ooni/probe-engine/pkg/experiment/example"
	"github.com/ooni/probe-engine/pkg/model"
)

func init() {
	AllExperiments["example"] = &Factory{
		build: func(config interface{}) model.ExperimentMeasurer {
			return example.NewExperimentMeasurer(
				*config.(*example.Config), "example",
			)
		},
		config: &example.Config{
			Message:   "Good day from the example experiment!",
			SleepTime: int64(time.Second),
		},
		interruptible: true,
		inputPolicy:   model.InputNone,
	}
}
