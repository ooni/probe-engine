package probeservices

import (
	"context"
	"fmt"

	"github.com/ooni/probe-engine/internal/httpx"
)

// FetchPsiphonConfig fetches psiphon config from authenticated OONI orchestra.
func (c Client) FetchPsiphonConfig(ctx context.Context) ([]byte, error) {
	_, auth, err := c.getCredsAndAuth()
	if err != nil {
		return nil, err
	}
	authorization := fmt.Sprintf("Bearer %s", auth.Token)
	return (httpx.Client{
		Authorization: authorization,
		BaseURL:       c.BaseURL,
		HTTPClient:    c.HTTPClient,
		Logger:        c.Logger,
		UserAgent:     c.UserAgent,
	}).FetchResource(ctx, "/api/v1/test-list/psiphon-config")
}
