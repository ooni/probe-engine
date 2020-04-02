package dnsoverhttps

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/miekg/dns"
)

func TestIntegrationDNSOverHTTPSSuccess(t *testing.T) {
	const queryURL = "https://cloudflare-dns.com/dns-query"
	transport := NewDNSOverHTTPS(
		http.DefaultClient, queryURL,
	)
	if transport.Network() != "doh" {
		t.Fatal("invalid network")
	}
	if transport.Address() != queryURL {
		t.Fatal("invalid address")
	}
	err := threeRounds(transport)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationDNSOverHTTPSNewRequestFailure(t *testing.T) {
	transport := NewDNSOverHTTPS(
		http.DefaultClient, "\t", // invalid URL
	)
	err := threeRounds(transport)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationDNSOverHTTPSClientDoFailure(t *testing.T) {
	transport := NewDNSOverHTTPS(
		http.DefaultClient, "https://cloudflare-dns.com/dns-query",
	)
	transport.clientDo = func(*http.Request) (*http.Response, error) {
		return nil, errors.New("mocked error")
	}
	err := threeRounds(transport)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationDNSOverHTTPSHTTPFailure(t *testing.T) {
	transport := NewDNSOverHTTPS(
		http.DefaultClient, "https://cloudflare-dns.com/dns-query",
	)
	transport.clientDo = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}, nil
	}
	err := threeRounds(transport)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationDNSOverHTTPSMissingHeader(t *testing.T) {
	transport := NewDNSOverHTTPS(
		http.DefaultClient, "https://cloudflare-dns.com/dns-query",
	)
	transport.clientDo = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}, nil
	}
	err := threeRounds(transport)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func threeRounds(transport *DNSOverHTTPS) error {
	err := roundTrip(transport, "ooni.io.")
	if err != nil {
		return err
	}
	err = roundTrip(transport, "slashdot.org.")
	if err != nil {
		return err
	}
	err = roundTrip(transport, "kernel.org.")
	if err != nil {
		return err
	}
	return nil
}

func roundTrip(transport *DNSOverHTTPS, domain string) error {
	query := new(dns.Msg)
	query.SetQuestion(domain, dns.TypeA)
	data, err := query.Pack()
	if err != nil {
		return err
	}
	data, err = transport.RoundTrip(context.Background(), data)
	if err != nil {
		return err
	}
	return query.Unpack(data)
}
