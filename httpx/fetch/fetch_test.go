package fetch_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/httpx/fetch"
	"github.com/ooni/probe-engine/httpx/httplog"
	"github.com/ooni/probe-engine/httpx/httptracex"
)

func TestFetchIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	client := &http.Client{
		Transport: &httptracex.Measurer{
			RoundTripper: http.DefaultTransport,
			Handler: &httplog.RoundTripLogger{
				Logger: log.Log,
			},
		},
	}
	data, err := (&fetch.Client{
		HTTPClient: client,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).Fetch(ctx, "http://facebook.com/robots.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) <= 0 {
		t.Fatal("Did not expect an empty resource")
	}
}

func TestFetchAndVerifyIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	client := &http.Client{
		Transport: &httptracex.Measurer{
			RoundTripper: http.DefaultTransport,
			Handler: &httplog.RoundTripLogger{
				Logger: log.Log,
			},
		},
	}
	data, err := (&fetch.Client{
		HTTPClient: client,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}).FetchAndVerify(
		ctx,
		"https://github.com/measurement-kit/generic-assets/releases/download/20190426155936/generic-assets-20190426155936.tar.gz",
		"34d8a9c8ab30c242469482dc280be832d8a06b4400f8927604dd361bf979b795",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) <= 0 {
		t.Fatal("Did not expect an empty resource")
	}
}
