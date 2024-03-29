//go:build ooni_libtor

package tunnel

//
// This file implements the ooni_libtor strategy of embedding tor. We manually
// compile tor and its dependencies and link against it.
//
// See https://github.com/ooni/probe/issues/2365 and
// https://github.com/ooni/probe/issues/2564.
//

import (
	"errors"
	"strings"

	"github.com/cretz/bine/tor"
	"github.com/ooni/probe-engine/pkg/libtor"
)

// getTorStartConf in this configuration returns a tor.StartConf
// configured to run the version of tor we embed as a library.
func getTorStartConf(config *Config, dataDir string, extraArgs []string) (*tor.StartConf, error) {
	creator, good := libtor.MaybeCreator()
	if !good {
		return nil, errors.New("no embedded tor")
	}
	config.logger().Infof("tunnel: tor: exec: <internal/libtor> %s %s",
		dataDir, strings.Join(extraArgs, " "))
	return &tor.StartConf{
		ProcessCreator:         creator,
		UseEmbeddedControlConn: true,
		DataDir:                dataDir,
		ExtraArgs:              extraArgs,
		NoHush:                 true,
	}, nil
}
