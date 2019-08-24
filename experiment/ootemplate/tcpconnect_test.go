package ootemplate_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/ootemplate"
)

func TestTCPConnectAsync(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	inputs := []string{
		"149.154.175.50:443", "149.154.167.51:443", "149.154.175.100:443",
		"149.154.167.91:443", "149.154.171.5:443",
	}
	for results := range ootemplate.TCPConnectAsync(ctx, log.Log, inputs...) {
		t.Logf("%+v", results)
	}
}
