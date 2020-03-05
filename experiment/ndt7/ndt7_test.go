package ndt7_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/experiment/ndt7"
	"github.com/ooni/probe-engine/model2"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model2.ExperimentMeasurer {
		return ndt7.NewExperimentMeasurer(ndt7.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
