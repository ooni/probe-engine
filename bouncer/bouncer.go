// Package bouncer contains a OONI bouncer client implementation.
//
// Specifically we implement v2.0.0 of the OONI bouncer specification defined
// in https://github.com/ooni/spec/blob/master/backends/bk-004-bouncer.md.
package bouncer

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/httpx/jsonapi"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
)

// Client is a client for the OONI bouncer API.
type Client struct {
	// BaseURL is the bouncer base URL.
	BaseURL string

	// HTTPClient is the HTTP client to use.
	HTTPClient *http.Client

	// Logger is the logger to use.
	Logger log.Logger

	// UserAgent is the user agent to use.
	UserAgent string
}

// GetCollectors queries the bouncer for collectors. Returns a list of
// entries on success; an error on failure.
func (c *Client) GetCollectors(
	ctx context.Context,
) (output []model.Service, err error) {
	err = (&jsonapi.Client{
		BaseURL:    c.BaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	}).Read(ctx, "/api/v1/collectors", &output)
	return
}

// GetTestHelpers is like GetCollectors but for test helpers.
func (c *Client) GetTestHelpers(
	ctx context.Context,
) (output map[string][]model.Service, err error) {
	err = (&jsonapi.Client{
		BaseURL:    c.BaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	}).Read(ctx, "/api/v1/test-helpers", &output)
	return
}
