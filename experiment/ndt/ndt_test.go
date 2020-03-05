package ndt_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/experiment/ndt"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model.ExperimentMeasurer {
		return ndt.NewExperimentMeasurer(ndt.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
