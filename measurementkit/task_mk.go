// +build !nomk

package measurementkit

import (
	"github.com/ooni/probe-engine/measurementkit/mkcgo"
)

func Start(settings []byte) (<-chan []byte, error) {
	return mkcgo.Start(settings)
}

func available() bool {
	return true
}
