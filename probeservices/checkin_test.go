package probeservices_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ooni/probe-engine/model"
)

func TestCheckInSuccess(t *testing.T) {
	client := newclient()
	client.BaseURL = "https://ams-pg-test.ooni.org"
	config := model.CheckInConfig{
		Charging:        true,
		OnWiFi:          true,
		Platform:        "android",
		ProbeASN:        "AS12353",
		ProbeCC:         "PT",
		RunType:         "timed",
		SoftwareVersion: "2.7.1",
		WebConnectivity: model.CategoryCodes{
			CategoryCodes: []string{"NEWS", "CULTR"},
		},
	}
	ctx := context.Background()
	result, err := client.CheckIn(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.WebConnectivity == nil {
		t.Fatal("got nil result or WebConnectivity")
	}
	if len(result.WebConnectivity.URLs) < 1 {
		t.Fatal("unexpected number of results")
	}
	for _, entry := range result.WebConnectivity.URLs {
		if entry.CategoryCode != "NEWS" && entry.CategoryCode != "CULTR" {
			t.Fatalf("unexpected category code: %+v", entry)
		}
	}
}

func TestCheckInFailure(t *testing.T) {
	client := newclient()
	client.BaseURL = "https://\t\t\t/" // cause test to fail
	config := model.CheckInConfig{
		Charging:        true,
		OnWiFi:          true,
		Platform:        "android",
		ProbeASN:        "AS12353",
		ProbeCC:         "PT",
		RunType:         "timed",
		SoftwareVersion: "2.7.1",
		WebConnectivity: model.CategoryCodes{
			CategoryCodes: []string{"NEWS", "CULTR"},
		},
	}
	ctx := context.Background()
	result, err := client.CheckIn(ctx, config)
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
	if result != nil {
		t.Fatal("results?!")
	}
}
