// +build !nomk

package dash_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/dash"
	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model.ExperimentMeasurer {
		return dash.NewExperimentMeasurer(dash.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
