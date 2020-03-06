package fbmessenger_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model.ExperimentMeasurer {
		return fbmessenger.NewExperimentMeasurer(fbmessenger.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
