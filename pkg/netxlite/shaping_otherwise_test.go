//go:build !shaping

package netxlite

import (
	"testing"

	"github.com/ooni/probe-engine/pkg/model/mocks"
)

func TestNewShapingDialer(t *testing.T) {
	in := &mocks.Dialer{}
	out := NewMaybeShapingDialer(in)
	if in != out {
		t.Fatal("expected to see the same pointer")
	}
}
