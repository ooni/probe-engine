package resources_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/resources"
)

func TestEnsure(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	client := resources.Client{
		HTTPClient: httpx.NewTracingProxyingClient(log.Log, nil),
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
		WorkDir:    "../testdata/",
	}
	err := client.Ensure(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// the second round should be idempotent
	err = client.Ensure(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
