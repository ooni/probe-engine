package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitCacheFailure(t *testing.T) {
	expected := errors.New("mocked error")
	var r resolver.Resolver = resolver.FakeResolver{
		Err: expected,
	}
	r = &resolver.CacheResolver{Resolver: r}
	addrs, err := r.LookupHost(context.Background(), "www.google.com")
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if addrs != nil {
		t.Fatal("expected nil addrs here")
	}
}

func TestUnitCacheHitSuccess(t *testing.T) {
	var r resolver.Resolver = resolver.FakeResolver{
		Err: errors.New("mocked error"),
	}
	cache := &resolver.CacheResolver{Resolver: r}
	cache.Set("dns.google.com", []string{"8.8.8.8"})
	addrs, err := cache.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0] != "8.8.8.8" {
		t.Fatal("not the result we expected")
	}
}

func TestUnitCacheMissSuccess(t *testing.T) {
	var r resolver.Resolver = resolver.FakeResolver{
		Result: []string{"8.8.8.8"},
	}
	r = &resolver.CacheResolver{Resolver: r}
	addrs, err := r.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0] != "8.8.8.8" {
		t.Fatal("not the result we expected")
	}
}
