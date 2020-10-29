package libminiooni_test

import (
	"testing"

	"github.com/ooni/probe-engine/libminiooni"
)

func TestIntegrationSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	libminiooni.MainWithConfiguration("example", libminiooni.Options{})
}
