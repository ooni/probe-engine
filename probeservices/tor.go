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
	client := c.Client
	client.Authorization = fmt.Sprintf("Bearer %s", auth.Token)
	err = client.GetJSON(ctx, "/api/v1/test-list/tor-targets", &result)
	return
}
