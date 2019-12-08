package engine

import (
	"context"
	"errors"

	"github.com/ooni/probe-engine/internal/orchestra/testlists/urls"
	"github.com/ooni/probe-engine/model"
)

// TestListsURLsConfig config config for test-lists/urls API.
type TestListsURLsConfig struct {
	BaseURL    string   // URL to use (empty means default)
	Categories []string // Categories to query for (empty means all)
	Limit      int      // Max number of URLs (<= 0 means no limit)
}

// AddCategory adds a category to the list of categories to query. Not
// adding any categories will query for URLs in all categories.
func (c *TestListsURLsConfig) AddCategory(s string) {
	c.Categories = append(c.Categories, s)
}

// TestListsURLsResult contains the results of calling the
// test-lists/urls OONI orchestra API.
type TestListsURLsResult struct {
	Result []model.URLInfo
}

// Count returns the number of returned URLs
func (r *TestListsURLsResult) Count() int64 {
	return int64(len(r.Result))
}

// At returns the URL at the given index or nil
func (r *TestListsURLsResult) At(idx int64) (out *model.URLInfo) {
	if idx >= 0 && idx < int64(len(r.Result)) {
		out = &r.Result[int(idx)]
	}
	return
}

// QueryTestListsURLs queries the test-lists/urls API.
func (sess *Session) QueryTestListsURLs(
	conf *TestListsURLsConfig,
) (*TestListsURLsResult, error) {
	if conf == nil {
		return nil, errors.New("QueryTestListURLs: passed nil config")
	}
	baseURL := "https://orchestrate.ooni.io"
	if conf.BaseURL != "" {
		baseURL = conf.BaseURL
	}
	result, err := urls.Query(context.Background(), urls.Config{
		BaseURL:           baseURL,
		CountryCode:       sess.ProbeCC(),
		EnabledCategories: conf.Categories,
		HTTPClient:        sess.session.HTTPDefaultClient,
		Limit:             conf.Limit,
		Logger:            sess.session.Logger,
		UserAgent:         sess.session.UserAgent(),
	})
	if err != nil {
		return nil, err
	}
	return &TestListsURLsResult{Result: result.Results}, nil
}
