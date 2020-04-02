package resolver_test

import (
	"net"
	"strings"
	"testing"

	"github.com/miekg/dns"
	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitDecoderUnpackError(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderNXDOMAIN(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(newErrorReply(t, dns.RcodeNameError))
	if err == nil || !strings.HasSuffix(err.Error(), "no such host") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderOtherError(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(newErrorReply(t, dns.RcodeRefused))
	if err == nil || !strings.HasSuffix(err.Error(), "query failed") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func newErrorReply(t *testing.T, code int) []byte {
	question := dns.Question{
		Name:   dns.Fqdn("x.org"),
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}
	query := new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = make([]dns.Question, 1)
	query.Question[0] = question
	reply := new(dns.Msg)
	reply.Compress = true
	reply.MsgHdr.RecursionAvailable = true
	reply.SetRcode(query, code)
	data, err := reply.Pack()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestUnitDecoderNoAddress(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(newSuccessReply(t, dns.TypeA))
	if err == nil || !strings.HasSuffix(err.Error(), "no response returned") {
		t.Fatal("not the error we expected")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}

func TestUnitDecoderDecodeA(t *testing.T) {
	d := resolver.MiekgDecoder{}
	data, err := d.Decode(newSuccessReply(t, dns.TypeA, "1.1.1.1", "8.8.8.8"))
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
	data, err := d.Decode(newSuccessReply(t, dns.TypeAAAA, "::1", "fe80::1"))
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

func newSuccessReply(t *testing.T, qtype uint16, ips ...string) []byte {
	question := dns.Question{
		Name:   dns.Fqdn("x.org"),
		Qtype:  qtype,
		Qclass: dns.ClassINET,
	}
	query := new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = make([]dns.Question, 1)
	query.Question[0] = question
	reply := new(dns.Msg)
	reply.Compress = true
	reply.MsgHdr.RecursionAvailable = true
	reply.SetReply(query)
	for _, ip := range ips {
		switch qtype {
		case dns.TypeA:
			reply.Answer = append(reply.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   dns.Fqdn("x.org"),
					Rrtype: qtype,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: net.ParseIP(ip),
			})
		case dns.TypeAAAA:
			reply.Answer = append(reply.Answer, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   dns.Fqdn("x.org"),
					Rrtype: qtype,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				AAAA: net.ParseIP(ip),
			})
		}
	}
	data, err := reply.Pack()
	if err != nil {
		t.Fatal(err)
	}
	return data
}
