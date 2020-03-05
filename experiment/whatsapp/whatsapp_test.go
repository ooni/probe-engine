package whatsapp_test

import (
	"testing"

	"github.com/ooni/probe-engine/experiment/mktesting"
	"github.com/ooni/probe-engine/experiment/whatsapp"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	err := mktesting.Run("", func() model.ExperimentMeasurer {
		return whatsapp.NewExperimentMeasurer(whatsapp.Config{})
	})
	if err != nil {
		t.Fatal(err)
	}
}
