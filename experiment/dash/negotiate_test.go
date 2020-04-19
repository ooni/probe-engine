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

func TestUnitNegotiateJSONMarshalError(t *testing.T) {
	expected := errors.New("mocked error")
	deps := FakeDeps{jsonMarshalErr: expected}
	result, err := negotiate(context.Background(), "", deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateNewHTTPRequestFailure(t *testing.T) {
	expected := errors.New("mocked error")
	deps := FakeDeps{newHTTPRequestErr: expected}
	result, err := negotiate(context.Background(), "", deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateHTTPClientDoFailure(t *testing.T) {
	expected := errors.New("mocked error")
	txp := FakeHTTPTransport{err: expected}
	deps := FakeDeps{httpTransport: txp, newHTTPRequestResult: &http.Request{
		Header: http.Header{},
		URL:    &url.URL{},
	}}
	result, err := negotiate(context.Background(), "", deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateInternalError(t *testing.T) {
	txp := FakeHTTPTransport{resp: &http.Response{StatusCode: 500}}
	deps := FakeDeps{httpTransport: txp, newHTTPRequestResult: &http.Request{
		Header: http.Header{},
		URL:    &url.URL{},
	}}
	result, err := negotiate(context.Background(), "", deps)
	if !errors.Is(err, errHTTPRequestFailed) {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateReadAllFailure(t *testing.T) {
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
	result, err := negotiate(context.Background(), "", deps)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateInvalidJSON(t *testing.T) {
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
	result, err := negotiate(context.Background(), "", deps)
	if err == nil || !strings.HasSuffix(err.Error(), "unexpected end of JSON input") {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateServerBusyFirstCase(t *testing.T) {
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
		readAllResult: []byte(`{"authorization": ""}`),
	}
	result, err := negotiate(context.Background(), "", deps)
	if !errors.Is(err, errServerBusy) {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateServerBusyThirdCase(t *testing.T) {
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
		readAllResult: []byte(`{}`),
	}
	result, err := negotiate(context.Background(), "", deps)
	if !errors.Is(err, errServerBusy) {
		t.Fatal("not the error we expected")
	}
	if result.Authorization != "" || result.Unchoked != 0 {
		t.Fatal("unexpected result")
	}
}

func TestUnitNegotiateSuccess(t *testing.T) {
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
		readAllResult: []byte(`{"authorization": "xx", "unchoked": 1}`),
	}
	result, err := negotiate(context.Background(), "", deps)
	if err != nil {
		t.Fatal(err)
	}
	if result.Authorization != "xx" || result.Unchoked != 1 {
		t.Fatal("invalid result")
	}
}
