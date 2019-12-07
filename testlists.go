package engine

import (
	"context"
	"errors"

	"github.com/ooni/probe-engine/internal/orchestra/testlists/urls"
)

// TestListsURLsConfig config config for test-lists/urls API.
type TestListsURLsConfig struct {
	baseURL    string
	categories []string
	limit      int
}

// AddCategory adds a category to the list of categories to query. Not
// adding any categories will query for URLs in all categories.
func (c *TestListsURLsConfig) AddCategory(s string) {
	c.categories = append(c.categories, s)
}

// SetLimit sets the maximum number of URLs to return. Not setting
// any limit, or setting a non-positive limit, cause the query to
// return all the available URLs without enforcing any limit.
func (c *TestListsURLsConfig) SetLimit(n int) {
	c.limit = n
}

// TestListsURLsResult contains the results of calling the
// test-lists/urls OONI orchestra API.
type TestListsURLsResult struct {
	result urls.Result
}

// Count returns the number of returned URLs
func (r *TestListsURLsResult) Count() int64 {
	return int64(len(r.result.Results))
}

// At returns the URL at the given index or nil
func (r *TestListsURLsResult) At(idx int64) TestListsURLInfo {
	if idx < 0 || idx >= int64(len(r.result.Results)) {
		return nil
	}
	return &urlinfo{u: r.result.Results[int(idx)]}
}

// QueryTestListsURLs queries the test-lists/urls API.
func (sess *Session) QueryTestListsURLs(
	conf *TestListsURLsConfig,
) (*TestListsURLsResult, error) {
	if conf == nil {
		return nil, errors.New("QueryTestListURLs: passed nil config")
	}
	baseURL := "https://orchestrate.ooni.io"
	if conf.baseURL != "" {
		baseURL = conf.baseURL
	}
	result, err := urls.Query(context.Background(), urls.Config{
		BaseURL:           baseURL,
		CountryCode:       sess.ProbeCC(),
		EnabledCategories: conf.categories,
		HTTPClient:        sess.session.HTTPDefaultClient,
		Limit:             conf.limit,
		Logger:            sess.session.Logger,
		UserAgent:         sess.session.UserAgent(),
	})
	if err != nil {
		return nil, err
	}
	return &TestListsURLsResult{result: result}, nil
}

// TestListsURLInfo contains info about URLs
type TestListsURLInfo interface {
	CategoryCode() string
	CountryCode() string
	URL() string
}

type urlinfo struct {
	u urls.URLInfo
}

func (u *urlinfo) URL() string {
	return u.u.URL
}

func (u *urlinfo) CategoryCode() string {
	return u.u.CategoryCode
}

func (u *urlinfo) CountryCode() string {
	return u.u.CountryCode
}
