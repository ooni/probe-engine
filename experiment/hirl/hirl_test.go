package hirl_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/model2"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model2.ExperimentMeasurer {
		return hirl.NewExperimentMeasurer(hirl.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
