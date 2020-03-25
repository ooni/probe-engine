package measurer_test

import (
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/measurer"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	results, err := measurer.Get("http://facebook.com", measurer.Config{Logger: log.Log})
	t.Log(err)
	for _, entry := range results.Resolutions {
		t.Logf("%+v", entry)
	}
	for _, entry := range results.Connects {
		t.Logf("%+v", entry)
	}
	for _, entry := range results.TLSHandshakes {
		t.Logf("%+v", entry)
	}
	for _, entry := range results.HTTPRoundTrips {
		t.Logf("%+v", entry)
	}
	for _, entry := range results.HTTPBodies {
		t.Logf("%+v", entry)
	}
}
