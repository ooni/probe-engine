package engineresolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/pkg/engineresolver"
	"github.com/ooni/probe-engine/pkg/kvstore"
)

func TestSessionResolverGood(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	reso := &engineresolver.Resolver{
		KVStore: &kvstore.Memory{},
	}
	defer reso.CloseIdleConnections()
	if reso.Network() != "sessionresolver" {
		t.Fatal("unexpected Network")
	}
	if reso.Address() != "" {
		t.Fatal("unexpected Address")
	}
	addrs, err := reso.LookupHost(context.Background(), "google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) < 1 {
		t.Fatal("expected some addrs here")
	}
}
