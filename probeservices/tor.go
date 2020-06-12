package probeservices

import (
	"context"
	"fmt"

	"github.com/ooni/probe-engine/model"
)

// FetchTorTargets returns the targets for the tor experiment.
func (c Client) FetchTorTargets(ctx context.Context) (result map[string]model.TorTarget, err error) {
	_, auth, err := c.getCredsAndAuth()
	if err != nil {
		return nil, err
	}
	authorization := fmt.Sprintf("Bearer %s", auth.Token)
	client := c.Client
	client.Authorization = authorization
	err = client.ReadJSON(ctx, "/api/v1/test-list/tor-targets", &result)
	return
}
