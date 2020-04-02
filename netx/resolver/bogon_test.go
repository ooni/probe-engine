package resolver_test

import (
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestIntegrationResolverIsBogon(t *testing.T) {
	if resolver.IsBogon("antani") != true {
		t.Fatal("unexpected result")
	}
	if resolver.IsBogon("127.0.0.1") != true {
		t.Fatal("unexpected result")
	}
	if resolver.IsBogon("1.1.1.1") != false {
		t.Fatal("unexpected result")
	}
	if resolver.IsBogon("10.0.1.1") != true {
		t.Fatal("unexpected result")
	}
}
