package handlers_test

import (
	"testing"

	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestIntegration(t *testing.T) {
	handlers.NoHandler.OnMeasurement(modelx.Measurement{})
	handlers.StdoutHandler.OnMeasurement(modelx.Measurement{})
}
