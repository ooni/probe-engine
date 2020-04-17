package dash

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestUnitCollectJSONMarshalError(t *testing.T) {
	expected := errors.New("mocked error")
	deps := FakeDeps{jsonMarshalErr: expected}
	err := collect(context.Background(), "", "", nil, deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitCollectNewHTTPRequestFailure(t *testing.T) {
	expected := errors.New("mocked error")
	deps := FakeDeps{newHTTPRequestErr: expected}
	err := collect(context.Background(), "", "", nil, deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitCollectHTTPClientDoFailure(t *testing.T) {
	expected := errors.New("mocked error")
	txp := FakeHTTPTransport{err: expected}
	deps := FakeDeps{httpTransport: txp, newHTTPRequestResult: &http.Request{
		Header: http.Header{},
		URL:    &url.URL{},
	}}
	err := collect(context.Background(), "", "", nil, deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitCollectInternalError(t *testing.T) {
	txp := FakeHTTPTransport{resp: &http.Response{StatusCode: 500}}
	deps := FakeDeps{httpTransport: txp, newHTTPRequestResult: &http.Request{
		Header: http.Header{},
		URL:    &url.URL{},
	}}
	err := collect(context.Background(), "", "", nil, deps)
	if !errors.Is(err, errHTTPRequestFailed) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitCollectReadAllFailure(t *testing.T) {
	expected := errors.New("mocked error")
	txp := FakeHTTPTransport{resp: &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		StatusCode: 200,
	}}
	deps := FakeDeps{
		httpTransport: txp,
		newHTTPRequestResult: &http.Request{
			Header: http.Header{},
			URL:    &url.URL{},
		},
		readAllErr: expected,
	}
	err := collect(context.Background(), "", "", nil, deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitCollectInvalidJSON(t *testing.T) {
	txp := FakeHTTPTransport{resp: &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		StatusCode: 200,
	}}
	deps := FakeDeps{
		httpTransport: txp,
		newHTTPRequestResult: &http.Request{
			Header: http.Header{},
			URL:    &url.URL{},
		},
		readAllResult: []byte("["),
	}
	err := collect(context.Background(), "", "", nil, deps)
	if err == nil || !strings.HasSuffix(err.Error(), "unexpected end of JSON input") {
		t.Fatal("not the error we expected")
	}
}

func TestUnitCollectSuccess(t *testing.T) {
	txp := FakeHTTPTransport{resp: &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		StatusCode: 200,
	}}
	deps := FakeDeps{
		httpTransport: txp,
		newHTTPRequestResult: &http.Request{
			Header: http.Header{},
			URL:    &url.URL{},
		},
		readAllResult: []byte("[]"),
	}
	err := collect(context.Background(), "", "", nil, deps)
	if err != nil {
		t.Fatal(err)
	}
}
