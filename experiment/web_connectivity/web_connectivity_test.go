package web_connectivity_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/experiment/web_connectivity"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("http://www.example.com", func() model.ExperimentMeasurer {
		return web_connectivity.NewExperimentMeasurer(web_connectivity.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
