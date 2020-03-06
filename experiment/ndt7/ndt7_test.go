package ndt7_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model.ExperimentMeasurer {
		return ndt7.NewExperimentMeasurer(ndt7.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
