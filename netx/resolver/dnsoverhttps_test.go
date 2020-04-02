package resolver_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitDNSOverHTTPSNewRequestFailure(t *testing.T) {
	const invalidURL = "\t"
	txp := resolver.NewDNSOverHTTPS(http.DefaultClient, invalidURL)
	data, err := txp.RoundTrip(context.Background(), nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverHTTPSClientDoFailure(t *testing.T) {
	txp := resolver.DNSOverHTTPS{
		Do: func(*http.Request) (*http.Response, error) {
			return nil, errors.New("mocked error")
		},
		URL: "https://cloudflare-dns.com/dns-query",
	}
	data, err := txp.RoundTrip(context.Background(), nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverHTTPSHTTPFailure(t *testing.T) {
	txp := resolver.DNSOverHTTPS{
		Do: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       ioutil.NopCloser(strings.NewReader("")),
			}, nil
		},
		URL: "https://cloudflare-dns.com/dns-query",
	}
	data, err := txp.RoundTrip(context.Background(), nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverHTTPSMissingContentType(t *testing.T) {
	txp := resolver.DNSOverHTTPS{
		Do: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("")),
			}, nil
		},
		URL: "https://cloudflare-dns.com/dns-query",
	}
	data, err := txp.RoundTrip(context.Background(), nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitDNSOverHTTPSSuccess(t *testing.T) {
	body := []byte("AAA")
	txp := resolver.DNSOverHTTPS{
		Do: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
				Header: http.Header{
					"Content-Type": []string{"application/dns-message"},
				},
			}, nil
		},
		URL: "https://cloudflare-dns.com/dns-query",
	}
	data, err := txp.RoundTrip(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, body) {
		t.Fatal("not the response we expected")
	}
}

func TestUnitDNSOverHTTPTransportOK(t *testing.T) {
	const queryURL = "https://cloudflare-dns.com/dns-query"
	txp := resolver.NewDNSOverHTTPS(http.DefaultClient, queryURL)
	if txp.Network() != "doh" {
		t.Fatal("invalid network")
	}
	if txp.RequiresPadding() != true {
		t.Fatal("should require padding")
	}
	if txp.Address() != queryURL {
		t.Fatal("invalid address")
	}
}
