package probeservices

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

type checkInResult struct {
	Tests model.CheckInInfo `json:"tests"`
	V     int               `json:"v"`
}

// CheckIn pobes ask for tests to be run or otherwise go back to sleep
// The config argument contains the mandatory settings.
// Returns the list of tests to run and the URLs, on success, or an explanatory error, in case of failure.
func (c Client) CheckIn(ctx context.Context, config model.CheckInConfig) (*model.CheckInInfo, error) {
	var response checkInResult
	if err := c.Client.PostJSON(ctx, "/api/v1/check-in", config, &response); err != nil {
		return nil, err
	}
	return &response.Tests, nil
}
