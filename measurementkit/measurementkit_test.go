package measurementkit_test

import (
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/measurementkit"
)

func TestTaskIntegrationNDT(t *testing.T) {
	if !measurementkit.IsAvailable() {
		t.Skip("Measurement Kit support not compiled in")
	}
	log.SetLevel(log.DebugLevel)
	settings := measurementkit.NewSettings(
		"Ndt", "ooniprobe-example", "0.1.0",
		"../testdata/ca-bundle.pem", "AS30722", "IT",
		"130.25.149.142", "Vodafone Italia S.p.A.",
	)
	ch, err := measurementkit.StartEx(settings, log.Log)
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		// Drain
	}
}
