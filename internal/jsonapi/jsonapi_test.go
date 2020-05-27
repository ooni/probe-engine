package jsonapi_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/netx/dialer"
)

func newClient() jsonapi.Client {
	return jsonapi.Client{
		BaseURL:    "https://httpbin.org",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	}
}

func TestNewRequestWithJSONBodyJSONMarshalFailure(t *testing.T) {
	client := newClient()
	req, err := client.NewRequest(
		context.Background(), "GET", "/", nil, make(chan interface{}),
	)
	if err == nil || !strings.HasPrefix(err.Error(), "json: unsupported type") {
		t.Fatal("not the error we expected")
	}
	if req != nil {
		t.Fatal("expected nil request here")
	}
}

func TestNewRequestWithJSONBodyNewRequestFailure(t *testing.T) {
	client := newClient()
	client.BaseURL = "\t\t\t" // cause URL parse error
	req, err := client.NewRequest(
		context.Background(), "GET", "/", nil, nil,
	)
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
	if req != nil {
		t.Fatal("expected nil request here")
	}
}

func TestNewRequestWithQuery(t *testing.T) {
	client := newClient()
	q := url.Values{}
	q.Add("antani", "mascetti")
	q.Add("melandri", "conte")
	req, err := client.NewRequest(
		context.Background(), "GET", "/", q, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.Query().Get("antani") != "mascetti" {
		t.Fatal("expected different query string here")
	}
	if req.URL.Query().Get("melandri") != "conte" {
		t.Fatal("expected different query string here")
	}
}

func TestNewRequestNewRequestFailure(t *testing.T) {
	client := newClient()
	req, err := client.NewRequest(
		context.Background(), "\t\t\t", "/", nil, nil,
	)
	if err == nil || !strings.HasPrefix(err.Error(), "net/http: invalid method") {
		t.Fatal("not the error we expected")
	}
	if req != nil {
		t.Fatal("expected nil request here")
	}
}

func TestNewRequestCloudfronting(t *testing.T) {
	client := newClient()
	client.Host = "www.x.org"
	req, err := client.NewRequest(
		context.Background(), "GET", "/", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.Host != client.Host {
		t.Fatal("expected different req.Host here")
	}
}

func TestNewRequestContentTypeIsSet(t *testing.T) {
	client := newClient()
	req, err := client.NewRequest(
		context.Background(), "GET", "/", nil, []string{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatal("expected different Content-Type here")
	}
}

func TestNewRequestAuthorizationHeader(t *testing.T) {
	client := newClient()
	client.Authorization = "deadbeef"
	req, err := client.NewRequest(
		context.Background(), "GET", "/", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != client.Authorization {
		t.Fatal("expected different Authorization here")
	}
}

func TestNewRequestUserAgentIsSet(t *testing.T) {
	client := newClient()
	req, err := client.NewRequest(
		context.Background(), "GET", "/", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("User-Agent") != client.UserAgent {
		t.Fatal("expected different User-Agent here")
	}
}

func TestNewRequestTunnelingIsPossible(t *testing.T) {
	client := newClient()
	client.ProxyURL = &url.URL{Scheme: "socks5", Host: "[::1]:54321"}
	req, err := client.NewRequest(
		context.Background(), "GET", "/", nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	cmp := cmp.Diff(dialer.ContextProxyURL(req.Context()), client.ProxyURL)
	if cmp != "" {
		t.Fatal(cmp)
	}
}

func TestClientDoClientDoFailure(t *testing.T) {
	expected := errors.New("mocked error")
	client := newClient()
	client.HTTPClient = &http.Client{Transport: jsonapi.FakeTransport{
		Err: expected,
	}}
	err := client.Do(&http.Request{URL: &url.URL{Scheme: "https", Host: "x.org"}}, nil)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestClientDoResponseNotSuccessful(t *testing.T) {
	client := newClient()
	client.HTTPClient = &http.Client{Transport: jsonapi.FakeTransport{
		Resp: &http.Response{
			StatusCode: 401,
			Body:       jsonapi.FakeBody{},
		},
	}}
	err := client.Do(&http.Request{URL: &url.URL{Scheme: "https", Host: "x.org"}}, nil)
	if err == nil || !strings.HasPrefix(err.Error(), "jsonapi: request failed") {
		t.Fatal("not the error we expected")
	}
}

func TestClientDoResponseReadingBodyError(t *testing.T) {
	expected := errors.New("mocked error")
	client := newClient()
	client.HTTPClient = &http.Client{Transport: jsonapi.FakeTransport{
		Resp: &http.Response{
			StatusCode: 200,
			Body: jsonapi.FakeBody{
				Err: expected,
			},
		},
	}}
	err := client.Do(&http.Request{URL: &url.URL{Scheme: "https", Host: "x.org"}}, nil)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestClientDoResponseIsNotJSON(t *testing.T) {
	client := newClient()
	client.HTTPClient = &http.Client{Transport: jsonapi.FakeTransport{
		Resp: &http.Response{
			StatusCode: 200,
			Body: jsonapi.FakeBody{
				Err: io.EOF,
			},
		},
	}}
	err := client.Do(&http.Request{URL: &url.URL{Scheme: "https", Host: "x.org"}}, nil)
	if err == nil || err.Error() != "unexpected end of JSON input" {
		t.Fatal("not the error we expected")
	}
}

type httpbinheaders struct {
	Headers map[string]string `json:"headers"`
}

func TestIntegrationReadSuccess(t *testing.T) {
	var headers httpbinheaders
	err := newClient().Read(context.Background(), "/headers", &headers)
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
	err := newClient().Create(context.Background(), "/post", &headers, &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.Data != `{"headers":{"Foo":"bar"}}` {
		t.Fatal(response.Data)
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
	err := newClient().Update(context.Background(), "/put", &headers, &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.Data != `{"headers":{"Foo":"bar"}}` {
		t.Fatal(response.Data)
	}
}

func TestUnitReadFailure(t *testing.T) {
	var headers httpbinheaders
	client := newClient()
	client.BaseURL = "\t\t\t\t"
	err := client.Read(context.Background(), "/headers", &headers)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestUnitCreateFailure(t *testing.T) {
	var headers httpbinheaders
	client := newClient()
	client.BaseURL = "\t\t\t\t"
	err := client.Create(context.Background(), "/headers", &headers, &headers)
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestUnitUpdateFailure(t *testing.T) {
	var headers httpbinheaders
	client := newClient()
	client.BaseURL = "\t\t\t\t"
	err := client.Update(context.Background(), "/headers", &headers, &headers)
	if err == nil {
		t.Fatal("expected an error here")
	}
}
