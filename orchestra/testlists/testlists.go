// Package testlists queries orchestra's test lists.
package testlists

import (
	"context"
	"fmt"
	"strings"
	"net/http"
	"net/url"

	"github.com/ooni/probe-engine/httpx/jsonapi"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/session"
)

const (
	// DefaultBaseURL is the default base URL
	DefaultBaseURL = "https://orchestrate.ooni.io"
)

// URLInfo contains the URL and the citizenlab category code for that URL
type URLInfo struct {
	// URL is the URL
	URL string `json:"url"`

	// CountryCode is the country code
	CountryCode string `json:"country_code"`

	// CategoryCode is the category code
	CategoryCode string `json:"category_code"`
}

type response struct {
	Results []URLInfo `json:"results"`
}

// Client is a client for the requesting test lists.
type Client struct {
	// BaseURL is the orchestra base URL.
	BaseURL string

	// HTTPClient is the HTTP client to use.
	HTTPClient *http.Client

	// Logger is the logger to use.
	Logger log.Logger

	// UserAgent is the user agent to use.
	UserAgent string

	// EnabledCategories is a list of category codes that are enabled
	EnabledCategories []string
}

// NewClient creates a new client in the context of the given session.
func NewClient(sess *session.Session) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: sess.HTTPDefaultClient,
		Logger:     sess.Logger,
		UserAgent:  sess.UserAgent(),
	}
}

// SetEnabledCategories configures the client category codes
func (c *Client) SetEnabledCategories(categories []string) error {
	c.EnabledCategories = categories
	return nil
}

// Do retrieves the test list for the specified country.
func (c *Client) Do(
	ctx context.Context, countryCode string, limit int,
) ([]URLInfo, error) {
	var resp response
	query := url.Values{}
	if countryCode != "" {
		query.Set("probe_cc", countryCode)
	}
	if (limit > 0) {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if (len(c.EnabledCategories) > 0) {
		query.Set("category_codes", strings.Join(c.EnabledCategories, ","))
	}
	err := (&jsonapi.Client{
		BaseURL:    c.BaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	}).ReadWithQuery(ctx, "/api/v1/urls", query, &resp)
	return resp.Results, err
}
