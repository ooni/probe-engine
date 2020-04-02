package resolver

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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
