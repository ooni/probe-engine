// Package urls queries orchestra test-lists/urls API
package urls

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
)

// Config contains configs for querying tests-lists/urls
type Config struct {
	BaseURL           string
	CountryCode       string
	EnabledCategories []string
	HTTPClient        *http.Client
	Limit             int64
	Logger            log.Logger
	UserAgent         string
}

// Result contains the result returned by tests-lists/urls
type Result struct {
	Results []model.URLInfo `json:"results"`
}

// Query retrieves the test list for the specified country.
func Query(ctx context.Context, config Config) (*Result, error) {
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
	var response Result
	err := (&jsonapi.Client{
		BaseURL:    config.BaseURL,
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		UserAgent:  config.UserAgent,
	}).ReadWithQuery(ctx, "/api/v1/test-list/urls", query, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
