package probeservices

import (
	"context"
	"fmt"
)

// FetchPsiphonConfig fetches psiphon config from authenticated OONI orchestra.
func (c Client) FetchPsiphonConfig(ctx context.Context) ([]byte, error) {
	_, auth, err := c.getCredsAndAuth()
	if err != nil {
		return nil, err
	}
	authorization := fmt.Sprintf("Bearer %s", auth.Token)
	client := c.Client
	client.Authorization = authorization
	return client.FetchResource(ctx, "/api/v1/test-list/psiphon-config")
}
