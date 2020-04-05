package httptransport_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/httptransport"
)

func TestUnitByteCounterFailure(t *testing.T) {
	counter := bytecounter.New()
	txp := httptransport.ByteCountingTransport{
		Counter: counter,
		RoundTripper: httptransport.MockableTransport{
			Err: io.EOF,
		},
	}
	client := &http.Client{Transport: txp}
	req, err := http.NewRequest(
		"POST", "https://www.google.com", strings.NewReader("AAAAAA"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "antani-browser/1.0.0")
	resp, err := client.Do(req)
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
	if counter.Sent.Load() != 68 {
		t.Fatal("expected around 68 bytes sent")
	}
	if counter.Received.Load() != 0 {
		t.Fatal("expected zero bytes received")
	}
}

func TestUnitByteCounterSuccess(t *testing.T) {
	counter := bytecounter.New()
	txp := httptransport.ByteCountingTransport{
		Counter: counter,
		RoundTripper: httptransport.MockableTransport{
			Resp: &http.Response{
				Body: ioutil.NopCloser(strings.NewReader("1234567")),
				Header: http.Header{
					"Server": []string{"antani/0.1.0"},
				},
				Status:     "200 OK",
				StatusCode: http.StatusOK,
			},
		},
	}
	client := &http.Client{Transport: txp}
	req, err := http.NewRequest(
		"POST", "https://www.google.com", strings.NewReader("AAAAAA"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "antani-browser/1.0.0")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if counter.Sent.Load() != 68 {
		t.Fatal("expected around 68 bytes sent")
	}
	if counter.Received.Load() != 37 {
		t.Fatal("expected zero around 37 bytes received")
	}
}
