package probeservices

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

type checkinResult struct {
	Tests model.CheckinInfo `json:"tests"`
	V     string            `json:"v"`
}

// CheckIn pobes ask for tests to be run or otherwise go back to sleep
// The config argument contains the mandatory settings.
// Returns the list of tests to run and the URLs, on success, or an explanatory error, in case of failure.
func (c Client) CheckIn(ctx context.Context, config model.CheckinConfig) (model.CheckinInfo, error) {

	var response checkinResult
	err := c.Client.PostJSON(ctx, "/api/v1/check-in", config, &response)
	if err != nil {
		//TOCHECK this will probably be nil but with "nil" it doesn't compile
		return response.Tests, err
	}
	return response.Tests, nil
}
