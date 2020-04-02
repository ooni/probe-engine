package resolver_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitDNSOverUDPDialFailure(t *testing.T) {
	mocked := errors.New("mocked error")
	const address = "9.9.9.9:53"
	txp := resolver.NewDNSOverUDP(resolver.FakeDialer{Err: mocked}, address)
	data, err := txp.RoundTrip(context.Background(), nil)
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverUDPSetDeadlineError(t *testing.T) {
	mocked := errors.New("mocked error")
	txp := resolver.NewDNSOverUDP(
		resolver.FakeDialer{
			Conn: resolver.FakeConn{
				SetDeadlineError: mocked,
			},
		}, "9.9.9.9:53",
	)
	data, err := txp.RoundTrip(context.Background(), nil)
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverUDPWriteFailure(t *testing.T) {
	mocked := errors.New("mocked error")
	txp := resolver.NewDNSOverUDP(
		resolver.FakeDialer{
			Conn: resolver.FakeConn{
				WriteError: mocked,
			},
		}, "9.9.9.9:53",
	)
	data, err := txp.RoundTrip(context.Background(), nil)
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverUDPReadFailure(t *testing.T) {
	mocked := errors.New("mocked error")
	txp := resolver.NewDNSOverUDP(
		resolver.FakeDialer{
			Conn: resolver.FakeConn{
				ReadError: mocked,
			},
		}, "9.9.9.9:53",
	)
	data, err := txp.RoundTrip(context.Background(), nil)
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverUDPReadSuccess(t *testing.T) {
	const expected = 17
	txp := resolver.NewDNSOverUDP(
		resolver.FakeDialer{
			Conn: resolver.FakeConn{ReadCount: expected},
		}, "9.9.9.9:53",
	)
	data, err := txp.RoundTrip(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != expected {
		t.Fatal("expected non nil data")
	}
}

func TestUnitDNSOverUDPTransportOK(t *testing.T) {
	const address = "9.9.9.9:53"
	txp := resolver.NewDNSOverUDP(&net.Dialer{}, address)
	if txp.RequiresPadding() != false {
		t.Fatal("invalid RequiresPadding")
	}
	if txp.Network() != "udp" {
		t.Fatal("invalid Network")
	}
	if txp.Address() != address {
		t.Fatal("invalid Address")
	}
}
