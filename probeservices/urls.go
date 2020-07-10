package probeservices

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

type urlListResult struct {
	Results []model.URLInfo `json:"results"`
}

// FetchURLList fetches the list of URLs used by WebConnectivity. The config
// argument contains the optional settings. Returns the list of URLs, on success,
// or an explanatory error, in case of failure.
func (c Client) FetchURLList(ctx context.Context, config model.URLListConfig) ([]model.URLInfo, error) {
	query := url.Values{}
	if config.CountryCode != "" {
		query.Set("country_code", config.CountryCode)
	}
	if config.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", config.Limit))
	}
	if len(config.Categories) > 0 {
		query.Set("category_codes", strings.Join(config.Categories, ","))
	}
	var response urlListResult
	err := c.Client.GetJSONWithQuery(ctx, "/api/v1/test-list/urls", query, &response)
	if err != nil {
		return nil, err
	}
	return response.Results, nil
}

// URLsConfig contains configs for querying tests-lists/urls
//
// This structure is deprecated and will be removed in the future.
type URLsConfig struct {
	BaseURL           string
	CountryCode       string
	EnabledCategories []string
	HTTPClient        *http.Client
	Limit             int64
	Logger            model.Logger
	UserAgent         string
}

// URLsResult contains the result returned by tests-lists/urls
//
// This structure is deprecated and will be removed in the future.
type URLsResult urlListResult

// URLsQuery retrieves the test list for the specified country.
//
// This function is deprecated and will be removed in the future.
func URLsQuery(ctx context.Context, config URLsConfig) (*URLsResult, error) {
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
	var response URLsResult
	err := (httpx.Client{
		BaseURL:    config.BaseURL,
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		UserAgent:  config.UserAgent,
	}).GetJSONWithQuery(ctx, "/api/v1/test-list/urls", query, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
