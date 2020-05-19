// Package probeservices contains code to contact OONI probe services.
//
// Specifically we implement v2.0.0 of the OONI bouncer specification defined
// in https://github.com/ooni/spec/blob/master/backends/bk-004-bouncer
//
// We additionally implement v2.0.0 of the OONI collector specification defined
// in https://github.com/ooni/spec/blob/master/backends/bk-003-collector.md.
package probeservices

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/model"
)

// Client is a client for the OONI probe services API.
type Client struct {
	// BaseURL is the probe services base URL.
	BaseURL string

	// HTTPClient is the HTTP client to use.
	HTTPClient *http.Client

	// Host allows to force a host header for cloudfronting.
	Host string

	// Logger is the logger to use.
	Logger model.Logger

	// ProxyURL allows to force a proxy URL to fallback to a tunnel.
	ProxyURL *url.URL

	// UserAgent is the user agent to use.
	UserAgent string
}

// GetCollectors queries the bouncer for collectors. Returns a list of
// entries on success; an error on failure.
func (c *Client) GetCollectors(ctx context.Context) (output []model.Service, err error) {
	err = (jsonapi.Client{
		BaseURL:    c.BaseURL,
		HTTPClient: c.HTTPClient,
		Host:       c.Host,
		Logger:     c.Logger,
		ProxyURL:   c.ProxyURL,
		UserAgent:  c.UserAgent,
	}).Read(ctx, "/api/v1/collectors", &output)
	return
}

// GetTestHelpers is like GetCollectors but for test helpers.
func (c *Client) GetTestHelpers(
	ctx context.Context) (output map[string][]model.Service, err error) {
	err = (jsonapi.Client{
		BaseURL:    c.BaseURL,
		HTTPClient: c.HTTPClient,
		Host:       c.Host,
		Logger:     c.Logger,
		ProxyURL:   c.ProxyURL,
		UserAgent:  c.UserAgent,
	}).Read(ctx, "/api/v1/test-helpers", &output)
	return
}
