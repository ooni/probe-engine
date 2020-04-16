package httptransport_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/httptransport"
)

func TestUnitLoggingFailure(t *testing.T) {
	txp := httptransport.LoggingTransport{
		Logger: log.Log,
		RoundTripper: httptransport.FakeTransport{
			Err: io.EOF,
		},
	}
	client := &http.Client{Transport: txp}
	resp, err := client.Get("https://www.google.com")
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
}

func TestUnitLoggingFailureWithNoHostHeader(t *testing.T) {
	txp := httptransport.LoggingTransport{
		Logger: log.Log,
		RoundTripper: httptransport.FakeTransport{
			Err: io.EOF,
		},
	}
	req := &http.Request{
		Header: http.Header{},
		URL: &url.URL{
			Scheme: "https",
			Host:   "www.google.com",
			Path:   "/",
		},
	}
	resp, err := txp.RoundTrip(req)
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
}

func TestUnitLoggingSuccess(t *testing.T) {
	txp := httptransport.LoggingTransport{
		Logger: log.Log,
		RoundTripper: httptransport.FakeTransport{
			Resp: &http.Response{
				Body: ioutil.NopCloser(strings.NewReader("")),
				Header: http.Header{
					"Server": []string{"antani/0.1.0"},
				},
				StatusCode: 200,
			},
		},
	}
	client := &http.Client{Transport: txp}
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
}
