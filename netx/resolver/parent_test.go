package resolver_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/resolver"
)

type emitterchecker struct {
	containsBogons  bool
	gotResolveStart bool
	gotResolveDone  bool
	mu              sync.Mutex
}

func (h *emitterchecker) OnMeasurement(m modelx.Measurement) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m.ResolveStart != nil {
		h.gotResolveStart = true
	}
	if m.ResolveDone != nil {
		h.gotResolveDone = true
		h.containsBogons = m.ResolveDone.ContainsBogons
	}
}

func TestParentLookupHost(t *testing.T) {
	client := resolver.NewParentResolver(resolver.SystemResolver{})
	handler := new(emitterchecker)
	ctx := modelx.WithMeasurementRoot(
		context.Background(), &modelx.MeasurementRoot{
			Beginning: time.Now(),
			Handler:   handler,
		})
	addrs, err := client.LookupHost(ctx, "www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	for _, addr := range addrs {
		t.Log(addr)
	}
	handler.mu.Lock()
	defer handler.mu.Unlock()
	if handler.gotResolveStart == false {
		t.Fatal("did not see resolve start event")
	}
	if handler.gotResolveDone == false {
		t.Fatal("did not see resolve done event")
	}
	if handler.containsBogons == true {
		t.Fatal("did not expect to see bogons here")
	}
}

func TestParentLookupHostBogonHardError(t *testing.T) {
	client := resolver.NewParentResolver(resolver.SystemResolver{})
	handler := new(emitterchecker)
	ctx := modelx.WithMeasurementRoot(
		context.Background(), &modelx.MeasurementRoot{
			Beginning:   time.Now(),
			ErrDNSBogon: modelx.ErrDNSBogon,
			Handler:     handler,
		})
	addrs, err := client.LookupHost(ctx, "localhost")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if err.Error() != modelx.FailureDNSBogonError {
		t.Fatal("not the error that we expected")
	}
	if addrs != nil {
		t.Fatal("expected nil addr here")
	}
	if handler.gotResolveDone == false {
		t.Fatal("did not get the ResolveDone event")
	}
	if handler.containsBogons == false {
		t.Fatal("expected acknowledgement of bogons")
	}
}

func TestParentLookupHostBogonAsWarning(t *testing.T) {
	client := resolver.NewParentResolver(resolver.SystemResolver{})
	handler := new(emitterchecker)
	ctx := modelx.WithMeasurementRoot(
		context.Background(), &modelx.MeasurementRoot{
			Beginning: time.Now(),
			Handler:   handler,
		})
	addrs, err := client.LookupHost(ctx, "localhost")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expected non-nil addr here")
	}
	if handler.gotResolveDone == false {
		t.Fatal("did not get the ResolveDone event")
	}
	if handler.containsBogons == false {
		t.Fatal("expected acknowledgement of bogons")
	}
}
