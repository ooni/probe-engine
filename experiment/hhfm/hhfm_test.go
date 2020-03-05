package hhfm_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/hhfm"
	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model.ExperimentMeasurer {
		return hhfm.NewExperimentMeasurer(hhfm.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
