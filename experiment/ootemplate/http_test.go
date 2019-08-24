package ootemplate_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/ootemplate"
)

func TestHTTPPerformMany(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	templates := []ootemplate.HTTPRequestTemplate{
		ootemplate.HTTPRequestTemplate{
			Method: "GET",
			URL:    "http://google.com",
		},
		ootemplate.HTTPRequestTemplate{
			Method: "GET",
			URL:    "http://kernel.org",
		},
	}
	roundtrips, err := ootemplate.HTTPPerformMany(ctx, log.Log, templates...)
	if err != nil {
		t.Fatal(err)
	}
	for _, roundtrip := range roundtrips {
		data, err := json.MarshalIndent(roundtrip, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", string(data))
	}
}
