//go:build !shaping

package netxlite

import (
	"github.com/ooni/probe-engine/pkg/model"
)

func newMaybeShapingDialer(dialer model.Dialer) model.Dialer {
	return dialer
}
