package libminiooni_test

import (
	"testing"

	"github.com/ooni/probe-engine/libminiooni"
)

func TestIntegrationSimple(t *testing.T) {
	libminiooni.MainWithConfiguration("example", libminiooni.Options{})
}
