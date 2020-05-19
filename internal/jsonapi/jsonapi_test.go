package jsonapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/dialer"
)

type httpbinheaders struct {
	Headers map[string]string `json:"headers"`
}

func makeclient() Client {
	return Client{
		BaseURL:    "https://httpbin.org",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	}
}

func TestIntegrationReadSuccess(t *testing.T) {
	var headers httpbinheaders
	err := makeclient().Read(context.Background(), "/headers", &headers)
	if err != nil {
		t.Fatal(err)
	}
	if headers.Headers["Host"] != "httpbin.org" {
		t.Fatal("unexpected Host header")
	}
	if headers.Headers["User-Agent"] != "miniooni/0.1.0-dev" {
		t.Fatal("unexpected Host header")
	}
}

func TestUnitReadFailure(t *testing.T) {
	var headers httpbinheaders
	client := makeclient()
	client.BaseURL = "\t\t\t\t"
	err := client.Read(context.Background(), "/headers", &headers)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

type httpbinpost struct {
	Data string `json:"data"`
}

func TestIntegrationCreateSuccess(t *testing.T) {
	headers := httpbinheaders{
		Headers: map[string]string{
			"Foo": "bar",
		},
	}
	var response httpbinpost
	err := makeclient().Create(context.Background(), "/post", &headers, &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.Data != `{"headers":{"Foo":"bar"}}` {
		t.Fatal(response.Data)
	}
}

func TestUnitCreateFailure(t *testing.T) {
	var headers httpbinheaders
	client := makeclient()
	client.BaseURL = "\t\t\t\t"
	err := client.Create(context.Background(), "/headers", &headers, &headers)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

type httpbinput struct {
	Data string `json:"data"`
}

func TestIntegrationUpdateSuccess(t *testing.T) {
	headers := httpbinheaders{
		Headers: map[string]string{
			"Foo": "bar",
		},
	}
	var response httpbinpost
	err := makeclient().Update(context.Background(), "/put", &headers, &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.Data != `{"headers":{"Foo":"bar"}}` {
		t.Fatal(response.Data)
	}
}
func TestUnitUpdateFailure(t *testing.T) {
	var headers httpbinheaders
	client := makeclient()
	client.BaseURL = "\t\t\t\t"
	err := client.Update(context.Background(), "/headers", &headers, &headers)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestUnitMakeRequestWithJSONBodyFailure(t *testing.T) {
	client := makeclient()
	ch := make(chan interface{})
	req, err := client.makeRequestWithJSONBody(
		context.Background(), "PUT", "/", nil, ch,
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if req != nil {
		t.Fatal("expected nil req here")
	}
}

func TestUnitMakeRequestWithQuery(t *testing.T) {
	client := makeclient()
	query := url.Values{}
	query.Add("Foo", "bar")
	req, err := client.makeRequest(
		context.Background(), "PUT", "/", query, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if req == nil {
		t.Fatal("expected non-nil req here")
	}
	if req.URL.Query().Get("Foo") != "bar" {
		t.Fatal("unexpected query")
	}
}

func TestUnitMakeRequestWithInvalidMethod(t *testing.T) {
	client := makeclient()
	req, err := client.makeRequest(
		context.Background(), "\t", "/", nil, nil,
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if req != nil {
		t.Fatal("expected nil req here")
	}
}

func TestUnitMakeRequestWithAuthorization(t *testing.T) {
	client := makeclient()
	client.Authorization = "foo"
	req, err := client.makeRequest(
		context.Background(), "PUT", "/", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if req == nil {
		t.Fatal("expected non-nil req here")
	}
	if req.Header.Get("Authorization") != "foo" {
		t.Fatal("unexpected authorization")
	}
}

func TestIntegrationDoBadRequest(t *testing.T) {
	client := makeclient()
	req, err := client.makeRequest(
		context.Background(), "GET", "/status/404", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.do(req, nil); err == nil {
		t.Fatal("expected an error here")
	}
}

func TestUnitDoxDoFailure(t *testing.T) {
	client := makeclient()
	req, err := client.makeRequest(
		context.Background(), "GET", "/status/404", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	expected := errors.New("mocked error")
	err = client.dox(
		func(req *http.Request) (*http.Response, error) {
			return nil, expected
		},
		ioutil.ReadAll,
		req, nil,
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitDoxReadallFailure(t *testing.T) {
	client := makeclient()
	req, err := client.makeRequest(
		context.Background(), "GET", "/status/404", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	expected := errors.New("mocked error")
	err = client.dox(
		func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		},
		func(r io.Reader) ([]byte, error) {
			return nil, expected
		},
		req, nil,
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitCloudFronting(t *testing.T) {
	client := makeclient()
	client.Host = "www.x.org"
	req, err := client.makeRequest(
		context.Background(), "GET", "/status/404", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	expected := errors.New("mocked error")
	err = client.dox(
		func(req *http.Request) (*http.Response, error) {
			if req.Host != "www.x.org" {
				return nil, errors.New("expected req.Host to be set")
			}
			return nil, expected
		}, nil, req, nil)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitProxyTunnel(t *testing.T) {
	client := makeclient()
	client.ProxyURL = &url.URL{Scheme: "socks5", Host: "[::1]:443"}
	req, err := client.makeRequest(context.Background(), "GET", "/status/404", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := errors.New("mocked error")
	err = client.dox(
		func(req *http.Request) (*http.Response, error) {
			url := dialer.ContextProxyURL(req.Context())
			if url == nil || url.Scheme != "socks5" || url.Host != "[::1]:443" {
				return nil, errors.New("expected a different context URL")
			}
			return nil, expected
		}, nil, req, nil)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}
