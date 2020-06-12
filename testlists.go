package engine

import (
	"context"
	"errors"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/probeservices"
)

// TestListsURLsConfig config config for test-lists/urls API.
//
// This structure is deprecated and will be removed in the future.
type TestListsURLsConfig struct {
	BaseURL    string   // URL to use (empty means default)
	Categories []string // Categories to query for (empty means all)
	Limit      int64    // Max number of URLs (<= 0 means no limit)
}

// AddCategory adds a category to the list of categories to query. Not
// adding any categories will query for URLs in all categories.
//
// This function is deprecated and will be removed in the future.
func (c *TestListsURLsConfig) AddCategory(s string) {
	c.Categories = append(c.Categories, s)
}

// TestListsURLsResult contains the results of calling the
// test-lists/urls OONI orchestra API.
//
// This structure is deprecated and will be removed in the future.
type TestListsURLsResult struct {
	Result []model.URLInfo
}

// Count returns the number of returned URLs
//
// This function is deprecated and will be removed in the future.
func (r *TestListsURLsResult) Count() int64 {
	return int64(len(r.Result))
}

// At returns the URL at the given index or nil
//
// This function is deprecated and will be removed in the future.
func (r *TestListsURLsResult) At(idx int64) (out *model.URLInfo) {
	if idx >= 0 && idx < int64(len(r.Result)) {
		out = &r.Result[int(idx)]
	}
	return
}

// QueryTestListsURLs queries the test-lists/urls API.
//
// This function is deprecated and will be removed in the future. Please
// create and use a new orchestra client using the session instead.
func (s *Session) QueryTestListsURLs(conf *TestListsURLsConfig) (*TestListsURLsResult, error) {
	if conf == nil {
		return nil, errors.New("QueryTestListURLs: passed nil config")
	}
	baseURL := "https://ps.ooni.io"
	if conf.BaseURL != "" {
		baseURL = conf.BaseURL
	}
	result, err := probeservices.URLsQuery(context.Background(), probeservices.URLsConfig{
		BaseURL:           baseURL,
		CountryCode:       s.ProbeCC(),
		EnabledCategories: conf.Categories,
		HTTPClient:        s.DefaultHTTPClient(),
		Limit:             conf.Limit,
		Logger:            s.logger,
		UserAgent:         s.UserAgent(),
	})
	if err != nil {
		return nil, err
	}
	return &TestListsURLsResult{Result: result.Results}, nil
}
