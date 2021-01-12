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
		Platform:        "string",
		ProbeASN:        "string",
		ProbeCC:         "string",
		RunType:         "string",
		SoftwareVersion: "string",
		WebConnectivity: model.CategoryCodes{
			CategoryCodes: []string{"NEWS", "CULTR"},
		},
	}
	ctx := context.Background()
	result, err := client.CheckIn(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if result.WebConnectivity == nil {
		t.Fatal("got nil structure")
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
		Platform:        "string",
		ProbeASN:        "string",
		ProbeCC:         "string",
		RunType:         "string",
		SoftwareVersion: "string",
		WebConnectivity: model.CategoryCodes{
			CategoryCodes: []string{"NEWS", "CULTR"},
		},
	}
	ctx := context.Background()
	result, err := client.CheckIn(ctx, config)
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
	if result.WebConnectivity == nil {
		t.Fatal("results?!")
	}
}
