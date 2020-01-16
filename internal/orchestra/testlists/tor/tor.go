// Package tor contains code to fetch targets for the tor experiment.
package tor

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

// Config contains settings.
type Config struct{}

var staticTestingTargets = map[string]model.TorTarget{
	"f372c264e9a470335d9ac79fe780847bda052aa3b6a9ee5ff497cb6501634f9f": model.TorTarget{
		Address: "38.229.1.78:80",
		Params: map[string][]string{
			"cert": []string{
				"Hmyfd2ev46gGY7NoVxA9ngrPF2zCZtzskRTzoWXbxNkzeVnGFPWmrTtILRyqCTjHR+s9dg",
			},
			"iat-mode": []string{"1"},
		},
		Protocol: "obfs4",
	},
	"66bb51cfeaa6f3fc2694438a49f9245a2c35b994989b3986a09474088a4ea119": model.TorTarget{
		Address:  "66.111.2.131:9030",
		Protocol: "dir_port",
	},
	"11b0b9ce802244f7c2ebf5df315f40fb509ff1e1b445a52a5f0ff1190fd6dd09": model.TorTarget{
		Address:  "66.111.2.131:9001",
		Protocol: "or_port",
	},
}

// Query retrieves the tor experiment targets. This function will either
// return a nonzero list of targets or an error.
func Query(ctx context.Context, config Config) (map[string]model.TorTarget, error) {
	// TODO(bassosimone): fetch targets from orchestra
	return staticTestingTargets, nil
}
