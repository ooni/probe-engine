package resolver_test

import (
	"context"
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"

	"github.com/miekg/dns"
	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitOONIGettingTransport(t *testing.T) {
	txp := resolver.NewDNSOverTLS(resolver.DialTLSContext, "8.8.8.8:853")
	r := resolver.NewOONIResolver(txp)
	if txp != r.Transport() {
		t.Fatal("not the transport we expected")
	}
}

func TestUnitOONIEncodeError(t *testing.T) {
	mocked := errors.New("mocked error")
	txp := resolver.NewDNSOverTLS(resolver.DialTLSContext, "8.8.8.8:853")
	r := resolver.OONI{Encoder: resolver.FakeEncoder{Err: mocked}, Txp: txp}
	addrs, err := r.LookupHost(context.Background(), "www.gogle.com")
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if addrs != nil {
		t.Fatal("expected nil address here")
	}
}

func TestUnitOONIRoundTripError(t *testing.T) {
	mocked := errors.New("mocked error")
	txp := resolver.FakeTransport{Err: mocked}
	r := resolver.NewOONIResolver(txp)
	addrs, err := r.LookupHost(context.Background(), "www.gogle.com")
	if !errors.Is(err, mocked) {
		t.Fatal("not the error we expected")
	}
	if addrs != nil {
		t.Fatal("expected nil address here")
	}
}

func TestUnitOONIWithEmptyReply(t *testing.T) {
	txp := resolver.FakeTransport{Data: resolver.GenReplySuccess(t, dns.TypeA)}
	r := resolver.NewOONIResolver(txp)
	addrs, err := r.LookupHost(context.Background(), "www.gogle.com")
	if err == nil || !strings.HasSuffix(err.Error(), "no response returned") {
		t.Fatal("not the error we expected")
	}
	if addrs != nil {
		t.Fatal("expected nil address here")
	}
}

func TestUnitOONIWithAReply(t *testing.T) {
	txp := resolver.FakeTransport{
		Data: resolver.GenReplySuccess(t, dns.TypeA, "8.8.8.8"),
	}
	r := resolver.NewOONIResolver(txp)
	addrs, err := r.LookupHost(context.Background(), "www.gogle.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0] != "8.8.8.8" {
		t.Fatal("not the result we expected")
	}
}

func TestUnitOONIWithAAAAReply(t *testing.T) {
	txp := resolver.FakeTransport{
		Data: resolver.GenReplySuccess(t, dns.TypeAAAA, "::1"),
	}
	r := resolver.NewOONIResolver(txp)
	addrs, err := r.LookupHost(context.Background(), "www.gogle.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0] != "::1" {
		t.Fatal("not the result we expected")
	}
}

func TestUnitOONIWithTimeout(t *testing.T) {
	txp := resolver.FakeTransport{
		Err: &net.OpError{Err: syscall.ETIMEDOUT, Op: "dial"},
	}
	r := resolver.NewOONIResolver(txp)
	addrs, err := r.LookupHost(context.Background(), "www.gogle.com")
	if !errors.Is(err, syscall.ETIMEDOUT) {
		t.Fatal("not the error we expected")
	}
	if addrs != nil {
		t.Fatal("expected nil address here")
	}
	if r.NumTimeouts.Load() <= 0 {
		t.Fatal("we didn't actually take the timeouts")
	}
}
