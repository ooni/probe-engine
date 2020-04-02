package resolver_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitDNSOverTCPTransportQueryTooLarge(t *testing.T) {
	const address = "9.9.9.9:53"
	txp := resolver.NewDNSOverTCP(new(net.Dialer).DialContext, address)
	reply, err := txp.RoundTrip(context.Background(), make([]byte, 1<<18))
	if err == nil {
		t.Fatal("expected an error here")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
}

func TestUnitDNSOverTCPTransportDialFailure(t *testing.T) {
	const address = "9.9.9.9:53"
	mocked := errors.New("mocked error")
	fakedialer := resolver.FakeDialer{Err: mocked}
	txp := resolver.NewDNSOverTCP(fakedialer.DialContext, address)
	reply, err := txp.RoundTrip(context.Background(), make([]byte, 1<<11))
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
}

func TestUnitDNSOverTCPTransportSetDealineFailure(t *testing.T) {
	const address = "9.9.9.9:53"
	mocked := errors.New("mocked error")
	fakedialer := resolver.FakeDialer{Conn: &resolver.FakeConn{
		SetDeadlineError: mocked,
	}}
	txp := resolver.NewDNSOverTCP(fakedialer.DialContext, address)
	reply, err := txp.RoundTrip(context.Background(), make([]byte, 1<<11))
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
}

func TestUnitDNSOverTCPTransportWriteFailure(t *testing.T) {
	const address = "9.9.9.9:53"
	mocked := errors.New("mocked error")
	fakedialer := resolver.FakeDialer{Conn: &resolver.FakeConn{
		WriteError: mocked,
	}}
	txp := resolver.NewDNSOverTCP(fakedialer.DialContext, address)
	reply, err := txp.RoundTrip(context.Background(), make([]byte, 1<<11))
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
}

func TestUnitDNSOverTCPTransportReadFailure(t *testing.T) {
	const address = "9.9.9.9:53"
	mocked := errors.New("mocked error")
	fakedialer := resolver.FakeDialer{Conn: &resolver.FakeConn{
		ReadError: mocked,
	}}
	txp := resolver.NewDNSOverTCP(fakedialer.DialContext, address)
	reply, err := txp.RoundTrip(context.Background(), make([]byte, 1<<11))
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
}

func TestUnitDNSOverTCPTransportSecondReadFailure(t *testing.T) {
	const address = "9.9.9.9:53"
	mocked := errors.New("mocked error")
	fakedialer := resolver.FakeDialer{Conn: &resolver.FakeConn{
		ReadError: mocked,
		ReadData:  []byte{byte(0), byte(2)},
	}}
	txp := resolver.NewDNSOverTCP(fakedialer.DialContext, address)
	reply, err := txp.RoundTrip(context.Background(), make([]byte, 1<<11))
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if reply != nil {
		t.Fatal("expected nil reply here")
	}
}

func TestUnitDNSOverTCPTransportAllGood(t *testing.T) {
	const address = "9.9.9.9:53"
	mocked := errors.New("mocked error")
	fakedialer := resolver.FakeDialer{Conn: &resolver.FakeConn{
		ReadError: mocked,
		ReadData:  []byte{byte(0), byte(1), byte(1)},
	}}
	txp := resolver.NewDNSOverTCP(fakedialer.DialContext, address)
	reply, err := txp.RoundTrip(context.Background(), make([]byte, 1<<11))
	if err != nil {
		t.Fatal(err)
	}
	if len(reply) != 1 || reply[0] != 1 {
		t.Fatal("not the response we expected")
	}
}

func TestUnitDNSOverTCPTransportOK(t *testing.T) {
	const address = "9.9.9.9:53"
	txp := resolver.NewDNSOverTCP(new(net.Dialer).DialContext, address)
	if txp.RequiresPadding() != false {
		t.Fatal("invalid RequiresPadding")
	}
	if txp.Network() != "tcp" {
		t.Fatal("invalid Network")
	}
	if txp.Address() != address {
		t.Fatal("invalid Address")
	}
}

func TestUnitDNSOverTLSTransportOK(t *testing.T) {
	const address = "9.9.9.9:853"
	txp := resolver.NewDNSOverTLS(resolver.DialTLSContext, address)
	if txp.RequiresPadding() != true {
		t.Fatal("invalid RequiresPadding")
	}
	if txp.Network() != "dot" {
		t.Fatal("invalid Network")
	}
	if txp.Address() != address {
		t.Fatal("invalid Address")
	}
}
