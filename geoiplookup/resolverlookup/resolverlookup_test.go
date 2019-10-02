package resolverlookup_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
)

func TestResolverLookupAll(t *testing.T) {
	addrs, err := resolverlookup.All(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) < 1 {
		t.Fatal("expected a non-empty slice")
	}
}

func TestResolverLookupFirstSuccess(t *testing.T) {
	addr, err := resolverlookup.First(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if addr == "" {
		t.Fatal("expected a non-empty string")
	}
}

type brokenHostLookupper struct{}

func (*brokenHostLookupper) LookupHost(
	ctx context.Context, host string,
) (addrs []string, err error) {
	return
}

func TestResolverLookupFirstFailure(t *testing.T) {
	resolver := &brokenHostLookupper{}
	addr, err := resolverlookup.First(context.Background(), resolver)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if addr != "" {
		t.Fatal("expected an empty address")
	}
}
