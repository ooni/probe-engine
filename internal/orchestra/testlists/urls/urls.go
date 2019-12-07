// Package urls queries orchestra test-lists/urls API
package urls

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ooni/probe-engine/httpx/jsonapi"
	"github.com/ooni/probe-engine/log"
)

// Config contains configs for querying tests-lists/urls
type Config struct {
	BaseURL           string
	CountryCode       string
	EnabledCategories []string
	HTTPClient        *http.Client
	Limit             int
	Logger            log.Logger
	UserAgent         string
}

// Result contains the result returned by tests-lists/urls
type Result struct {
	Results []URLInfo `json:"results"`
}

// URLInfo contains the URL and the citizenlab category code for that URL
type URLInfo struct {
	CategoryCode string `json:"category_code"`
	CountryCode  string `json:"country_code"`
	URL          string `json:"url"`
}

// Query retrieves the test list for the specified country.
func Query(ctx context.Context, config Config) (response Result, err error) {
	query := url.Values{}
	if config.CountryCode != "" {
		query.Set("probe_cc", config.CountryCode)
	}
	if config.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", config.Limit))
	}
	if len(config.EnabledCategories) > 0 {
		query.Set("category_codes", strings.Join(config.EnabledCategories, ","))
	}
	err = (&jsonapi.Client{
		BaseURL:    config.BaseURL,
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		UserAgent:  config.UserAgent,
	}).ReadWithQuery(ctx, "/api/v1/test-list/urls", query, &response)
	return
}
