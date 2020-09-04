package handlers_test

import (
	"testing"

	"github.com/ooni/probe-engine/legacy/netx/handlers"
	"github.com/ooni/probe-engine/legacy/netx/modelx"
)

func TestIntegration(t *testing.T) {
	handlers.NoHandler.OnMeasurement(modelx.Measurement{})
	handlers.StdoutHandler.OnMeasurement(modelx.Measurement{})
	saver := handlers.SavingHandler{}
	saver.OnMeasurement(modelx.Measurement{})
	events := saver.Read()
	if len(events) != 1 {
		t.Fatal("invalid number of events")
	}
}
