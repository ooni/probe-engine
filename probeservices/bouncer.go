package probeservices

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

// GetTestHelpers is like GetCollectors but for test helpers.
func (c Client) GetTestHelpers(
	ctx context.Context) (output map[string][]model.Service, err error) {
	err = c.Client.GetJSON(ctx, "/api/v1/test-helpers", &output)
	return
}
