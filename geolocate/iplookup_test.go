package geolocate_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/geolocate"
	"github.com/ooni/probe-engine/model"
)

func TestIPLookupIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ip, err := (&geolocate.IPLookupClient{
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

func TestIPLookupAllFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel to cause Do() to fail
	log.SetLevel(log.DebugLevel)
	ip, err := (&geolocate.IPLookupClient{
		HTTPClient: http.DefaultClient,
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

func TestIPLookupInvalidIP(t *testing.T) {
	ctx := context.Background()
	log.SetLevel(log.DebugLevel)
	ip, err := (&geolocate.IPLookupClient{
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).DoWithCustomFunc(ctx, geolocate.InvalidIPLookup)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if ip != model.DefaultProbeIP {
		t.Fatal("expected the default IP here")
	}
}
