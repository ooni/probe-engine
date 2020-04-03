package resolver_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/resolver"
)

func TestEmittingTransportSuccess(t *testing.T) {
	ctx := context.Background()
	ctx = dialid.WithDialID(ctx)
	handler := &resolver.SavingHandler{}
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler:   handler,
	}
	ctx = modelx.WithMeasurementRoot(ctx, root)
	txp := resolver.Emitter{RoundTripper: resolver.FakeTransport{
		Data: resolver.GenReplySuccess(t, dns.TypeA, "8.8.8.8"),
	}}
	e := resolver.MiekgEncoder{}
	querydata, err := e.Encode("www.google.com", dns.TypeAAAA, true)
	if err != nil {
		t.Fatal(err)
	}
	replydata, err := txp.RoundTrip(ctx, querydata)
	if err != nil {
		t.Fatal(err)
	}
	events := handler.Read()
	if len(events) != 2 {
		t.Fatal("unexpected number of events")
	}
	if events[0].DNSQuery == nil {
		t.Fatal("missing DNSQuery field")
	}
	if !bytes.Equal(events[0].DNSQuery.Data, querydata) {
		t.Fatal("invalid query data")
	}
	if events[0].DNSQuery.DialID == 0 {
		t.Fatal("invalid query DialID")
	}
	if events[0].DNSQuery.DurationSinceBeginning <= 0 {
		t.Fatal("invalid duration since beginning")
	}
	if events[1].DNSReply == nil {
		t.Fatal("missing DNSReply field")
	}
	if !bytes.Equal(events[1].DNSReply.Data, replydata) {
		t.Fatal("missing reply data")
	}
	if events[1].DNSReply.DialID != 1 {
		t.Fatal("invalid query DialID")
	}
	if events[1].DNSReply.DurationSinceBeginning <= 0 {
		t.Fatal("invalid duration since beginning")
	}
}

func TestEmittingTransportFailure(t *testing.T) {
	ctx := context.Background()
	ctx = dialid.WithDialID(ctx)
	handler := &resolver.SavingHandler{}
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler:   handler,
	}
	ctx = modelx.WithMeasurementRoot(ctx, root)
	mocked := errors.New("mocked error")
	txp := resolver.Emitter{RoundTripper: resolver.FakeTransport{
		Err: mocked,
	}}
	e := resolver.MiekgEncoder{}
	querydata, err := e.Encode("www.google.com", dns.TypeAAAA, true)
	if err != nil {
		t.Fatal(err)
	}
	replydata, err := txp.RoundTrip(ctx, querydata)
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if replydata != nil {
		t.Fatal("expected nil replydata")
	}
	events := handler.Read()
	if len(events) != 1 {
		t.Fatal("unexpected number of events")
	}
	if events[0].DNSQuery == nil {
		t.Fatal("missing DNSQuery field")
	}
	if !bytes.Equal(events[0].DNSQuery.Data, querydata) {
		t.Fatal("invalid query data")
	}
	if events[0].DNSQuery.DialID == 0 {
		t.Fatal("invalid query DialID")
	}
	if events[0].DNSQuery.DurationSinceBeginning <= 0 {
		t.Fatal("invalid duration since beginning")
	}
}
