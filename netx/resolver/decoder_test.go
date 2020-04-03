package resolver_test

import (
	"strings"
	"testing"

	"github.com/miekg/dns"
	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitDecoderUnpackError(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(dns.TypeA, nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderNXDOMAIN(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(dns.TypeA, resolver.GenReplyError(t, dns.RcodeNameError))
	if err == nil || !strings.HasSuffix(err.Error(), "no such host") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderOtherError(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(dns.TypeA, resolver.GenReplyError(t, dns.RcodeRefused))
	if err == nil || !strings.HasSuffix(err.Error(), "query failed") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderNoAddress(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(dns.TypeA, resolver.GenReplySuccess(t, dns.TypeA))
	if err == nil || !strings.HasSuffix(err.Error(), "no response returned") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderDecodeA(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(
		dns.TypeA, resolver.GenReplySuccess(t, dns.TypeA, "1.1.1.1", "8.8.8.8"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 2 {
		t.Fatal("expected two entries here")
	}
	if data[0] != "1.1.1.1" {
		t.Fatal("invalid first IPv4 entry")
	}
	if data[1] != "8.8.8.8" {
		t.Fatal("invalid second IPv4 entry")
	}
}

func TestUnitDecoderDecodeAAAA(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(
		dns.TypeAAAA, resolver.GenReplySuccess(t, dns.TypeAAAA, "::1", "fe80::1"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 2 {
		t.Fatal("expected two entries here")
	}
	if data[0] != "::1" {
		t.Fatal("invalid first IPv6 entry")
	}
	if data[1] != "fe80::1" {
		t.Fatal("invalid second IPv6 entry")
	}
}

func TestUnitDecoderUnexpectedAReply(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(
		dns.TypeA, resolver.GenReplySuccess(t, dns.TypeAAAA, "::1", "fe80::1"))
	if err == nil || !strings.HasSuffix(err.Error(), "no response returned") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderUnexpectedAAAAReply(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(
		dns.TypeAAAA, resolver.GenReplySuccess(t, dns.TypeA, "1.1.1.1", "8.8.4.4."))
	if err == nil || !strings.HasSuffix(err.Error(), "no response returned") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}
