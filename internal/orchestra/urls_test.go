package orchestra_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra"
)

func TestURLsSuccess(t *testing.T) {
	config := orchestra.URLsConfig{
		BaseURL:           "https://ps.ooni.io",
		CountryCode:       "IT",
		EnabledCategories: []string{"NEWS", "CULTR"},
		HTTPClient:        http.DefaultClient,
		Limit:             17,
		Logger:            log.Log,
		UserAgent:         "ooniprobe-engine/v0.1.0-dev",
	}
	ctx := context.Background()
	result, err := orchestra.URLsQuery(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) < 1 {
		t.Fatal("no results")
	}
}

func TestURLsFailure(t *testing.T) {
	config := orchestra.URLsConfig{
		BaseURL:           "\t\t\t",
		CountryCode:       "IT",
		EnabledCategories: []string{"NEWS", "CULTR"},
		HTTPClient:        http.DefaultClient,
		Limit:             17,
		Logger:            log.Log,
		UserAgent:         "ooniprobe-engine/v0.1.0-dev",
	}
	ctx := context.Background()
	result, err := orchestra.URLsQuery(ctx, config)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if result != nil {
		t.Fatal("expected nil result here")
	}
}
