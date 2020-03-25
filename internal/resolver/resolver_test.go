package resolver_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/resolver"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	var r resolver.Resolver
	r = resolver.Base()
	r = resolver.ErrWrapper{Resolver: r}
	saver := &resolver.EventsSaver{Resolver: r}
	r = saver
	r = resolver.LoggingResolver{Resolver: r, Logger: log.Log}
	addrs, err := r.LookupHost(context.Background(), "www.facebook.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addrs)
	for _, ev := range saver.ReadEvents() {
		t.Logf("%+v", ev)
	}
}
