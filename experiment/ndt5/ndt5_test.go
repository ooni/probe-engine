package ndt5_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/experiment/ndt5"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model.ExperimentMeasurer {
		return ndt5.NewExperimentMeasurer(ndt5.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
