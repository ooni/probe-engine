package iplookup_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/geoiplookup/iplookup/invalid"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/model"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ip, err := (&iplookup.Client{
		HTTPClient: httpx.NewTracingProxyingClient(log.Log, nil, nil),
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

func TestAllFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel to cause Do() to fail
	log.SetLevel(log.DebugLevel)
	ip, err := (&iplookup.Client{
		HTTPClient: httpx.NewTracingProxyingClient(log.Log, nil, nil),
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).Do(ctx)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if ip != model.DefaultProbeIP {
		t.Fatal("expected the default IP here")
	}
}

func TestInvalidIP(t *testing.T) {
	ctx := context.Background()
	log.SetLevel(log.DebugLevel)
	ip, err := (&iplookup.Client{
		HTTPClient: httpx.NewTracingProxyingClient(log.Log, nil, nil),
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).DoWithCustomFunc(ctx, invalid.Do)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if ip != model.DefaultProbeIP {
		t.Fatal("expected the default IP here")
	}
}
