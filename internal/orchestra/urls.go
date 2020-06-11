package orchestra

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

// URLsConfig contains configs for querying tests-lists/urls
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
type URLsResult struct {
	Results []model.URLInfo `json:"results"`
}

// URLsQuery retrieves the test list for the specified country.
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
	err := (&httpx.Client{
		BaseURL:    config.BaseURL,
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		UserAgent:  config.UserAgent,
	}).ReadJSONWithQuery(ctx, "/api/v1/test-list/urls", query, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
