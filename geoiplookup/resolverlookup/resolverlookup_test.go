package resolverlookup_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
)

func TestResolverLookup(t *testing.T) {
	addrs, err := resolverlookup.Do(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, addr := range addrs {
		t.Log(addr)
	}
}
