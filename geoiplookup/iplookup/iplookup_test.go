package iplookup_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/internal/httpx"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ip, err := (&iplookup.Client{
		HTTPClient: httpx.NewTracingProxyingClient(log.Log, nil),
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}
