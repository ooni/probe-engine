package mlablocate_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/mlablocate"
)

func TestIntegrationWithoutProxy(t *testing.T) {
	client := mlablocate.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
	)
	fqdn, err := client.Query(context.Background(), "ndt7")
	if err != nil {
		t.Fatal(err)
	}
	if fqdn == "" {
		t.Fatal("unexpected empty fqdn")
	}
	t.Log(fqdn)
}

func TestIntegration404Response(t *testing.T) {
	client := mlablocate.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
	)
	fqdn, err := client.Query(context.Background(), "nonexistent")
	if err == nil || !strings.Contains(err.Error(), "mlablocate: non-200 status code") {
		t.Fatal("not the error we expected")
	}
	if fqdn != "" {
		t.Fatal("expected empty fqdn")
	}
}

func TestUnitNewRequestFailure(t *testing.T) {
	client := mlablocate.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
	)
	client.Hostname = "\t"
	fqdn, err := client.Query(context.Background(), "nonexistent")
	if err == nil || !strings.Contains(err.Error(), "invalid URL escape") {
		t.Fatal("not the error we expected")
	}
	if fqdn != "" {
		t.Fatal("expected empty fqdn")
	}
}

func TestUnitHTTPClientDoFailure(t *testing.T) {
	client := mlablocate.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
	)
	expected := errors.New("mocked error")
	client.HTTPClient = &http.Client{
		Transport: &roundTripFails{Error: expected},
	}
	fqdn, err := client.Query(context.Background(), "nonexistent")
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if fqdn != "" {
		t.Fatal("expected empty fqdn")
	}
}

type roundTripFails struct {
	Error error
}

func (txp *roundTripFails) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, txp.Error
}

func TestUnitCannotReadBody(t *testing.T) {
	client := mlablocate.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
	)
	expected := errors.New("mocked error")
	client.HTTPClient = &http.Client{
		Transport: &readingBodyFails{Error: expected},
	}
	fqdn, err := client.Query(context.Background(), "nonexistent")
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if fqdn != "" {
		t.Fatal("expected empty fqdn")
	}
}

type readingBodyFails struct {
	Error error
}

func (txp *readingBodyFails) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       &readingBodyFailsBody{Error: txp.Error},
	}, nil
}

type readingBodyFailsBody struct {
	Error error
}

func (b *readingBodyFailsBody) Read(p []byte) (int, error) {
	return 0, b.Error
}

func (b *readingBodyFailsBody) Close() error {
	return nil
}

func TestUnitInvalidJSON(t *testing.T) {
	client := mlablocate.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
	)
	client.HTTPClient = &http.Client{
		Transport: &invalidJSON{},
	}
	fqdn, err := client.Query(context.Background(), "nonexistent")
	if err == nil || !strings.Contains(err.Error(), "unexpected end of JSON input") {
		t.Fatal("not the error we expected")
	}
	if fqdn != "" {
		t.Fatal("expected empty fqdn")
	}
}

type invalidJSON struct{}

func (txp *invalidJSON) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       &invalidJSONBody{},
	}, nil
}

type invalidJSONBody struct{}

func (b *invalidJSONBody) Read(p []byte) (int, error) {
	if len(p) < 1 {
		return 0, errors.New("slice too short")
	}
	p[0] = '{'
	return 1, io.EOF
}

func (b *invalidJSONBody) Close() error {
	return nil
}

func TestUnitEmptyFQDN(t *testing.T) {
	client := mlablocate.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
	)
	client.HTTPClient = &http.Client{
		Transport: &emptyFQDN{},
	}
	fqdn, err := client.Query(context.Background(), "nonexistent")
	if err == nil || !strings.HasSuffix(err.Error(), "returned empty FQDN") {
		t.Fatal("not the error we expected")
	}
	if fqdn != "" {
		t.Fatal("expected empty fqdn")
	}
}

type emptyFQDN struct{}

func (txp *emptyFQDN) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       &emptyFQDNBody{},
	}, nil
}

type emptyFQDNBody struct{}

func (b *emptyFQDNBody) Read(p []byte) (int, error) {
	return copy(p, []byte(`{"fqdn":""}`)), io.EOF
}

func (b *emptyFQDNBody) Close() error {
	return nil
}
