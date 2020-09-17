package probeservices_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/probeservices"
)

func TestFetchURLListSuccess(t *testing.T) {
	client := newclient()
	client.BaseURL = "https://ams-pg.ooni.org"
	config := model.URLListConfig{
		Categories:  []string{"NEWS", "CULTR"},
		CountryCode: "IT",
		Limit:       17,
	}
	ctx := context.Background()
	result, err := client.FetchURLList(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 17 {
		t.Fatal("unexpected number of results")
	}
	for _, entry := range result {
		if entry.CategoryCode != "NEWS" && entry.CategoryCode != "CULTR" {
			t.Fatal("unexpected category code")
		}
	}
}

func TestFetchURLListFailure(t *testing.T) {
	client := newclient()
	client.BaseURL = "https://\t\t\t/" // cause test to fail
	config := model.URLListConfig{
		Categories:  []string{"NEWS", "CULTR"},
		CountryCode: "IT",
		Limit:       17,
	}
	ctx := context.Background()
	result, err := client.FetchURLList(ctx, config)
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
	if len(result) != 0 {
		t.Fatal("results?!")
	}
}

func TestURLsSuccess(t *testing.T) {
	config := probeservices.URLsConfig{
		BaseURL:           "https://ps1.ooni.io",
		CountryCode:       "IT",
		EnabledCategories: []string{"NEWS", "CULTR"},
		HTTPClient:        http.DefaultClient,
		Limit:             17,
		Logger:            log.Log,
		UserAgent:         "ooniprobe-engine/v0.1.0-dev",
	}
	ctx := context.Background()
	result, err := probeservices.URLsQuery(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) < 1 {
		t.Fatal("no results")
	}
}

func TestURLsFailure(t *testing.T) {
	config := probeservices.URLsConfig{
		BaseURL:           "\t\t\t",
		CountryCode:       "IT",
		EnabledCategories: []string{"NEWS", "CULTR"},
		HTTPClient:        http.DefaultClient,
		Limit:             17,
		Logger:            log.Log,
		UserAgent:         "ooniprobe-engine/v0.1.0-dev",
	}
	ctx := context.Background()
	result, err := probeservices.URLsQuery(ctx, config)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if result != nil {
		t.Fatal("expected nil result here")
	}
}
