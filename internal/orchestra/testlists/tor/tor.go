// Package tor contains code to fetch targets for the tor experiment.
package tor

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

// Config contains settings.
type Config struct{}

var staticTestingTargets = []model.TorTarget{
	// TODO(bassosimone): this is a public working bridge we have found
	// with @hellais. We should ask @phw whether there is some obfs4 bridge
	// dedicated to integration testing that we should use instead.
	model.TorTarget{
		Address: "109.105.109.165:10527",
		Params: map[string][]string{
			"cert": []string{
				"Bvg/itxeL4TWKLP6N1MaQzSOC6tcRIBv6q57DYAZc3b2AzuM+/TfB7mqTFEfXILCjEwzVA",
			},
			"iat-mode": []string{"1"},
		},
		Protocol: "obfs4",
	},
	model.TorTarget{
		Address:  "66.111.2.131:9030",
		Protocol: "dir_port",
	},
	model.TorTarget{
		Address:  "66.111.2.131:9001",
		Protocol: "or_port",
	},
}

// Query retrieves the tor experiment targets. This function will either
// return a nonzero list of targets or an error.
func Query(ctx context.Context, config Config) ([]model.TorTarget, error) {
	// TODO(bassosimone): fetch targets from orchestra
	return staticTestingTargets, nil
}
