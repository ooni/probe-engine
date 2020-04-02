package parentresolver

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/resolver/systemresolver"
)

func TestLookupAddr(t *testing.T) {
	client := NewParentResolver(new(net.Resolver))
	names, err := client.LookupAddr(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if names == nil {
		t.Fatal("expected non-nil result here")
	}
}

func TestLookupCNAME(t *testing.T) {
	client := NewParentResolver(new(net.Resolver))
	cname, err := client.LookupCNAME(context.Background(), "www.ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if cname == "" {
		t.Fatal("expected non-empty result here")
	}
}

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

func TestLookupHost(t *testing.T) {
	client := NewParentResolver(systemresolver.NewSystemResolver(new(net.Resolver)))
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

func TestLookupHostBogonHardError(t *testing.T) {
	client := NewParentResolver(systemresolver.NewSystemResolver(new(net.Resolver)))
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

func TestLookupHostBogonAsWarning(t *testing.T) {
	client := NewParentResolver(systemresolver.NewSystemResolver(new(net.Resolver)))
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

func TestLookupMX(t *testing.T) {
	client := NewParentResolver(new(net.Resolver))
	records, err := client.LookupMX(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if records == nil {
		t.Fatal("expected non-nil result here")
	}
}

func TestLookupNS(t *testing.T) {
	client := NewParentResolver(new(net.Resolver))
	records, err := client.LookupNS(context.Background(), "ooni.io")
	if err != nil {
		t.Fatal(err)
	}
	if records == nil {
		t.Fatal("expected non-nil result here")
	}
}
