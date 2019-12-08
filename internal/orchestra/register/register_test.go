package register

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
)

func TestIntegrationSuccess(t *testing.T) {
	result, err := Do(context.Background(), Config{
		BaseURL:    "https://ps-test.ooni.io",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		Metadata: metadata.Metadata{
			Platform:        "linux",
			ProbeASN:        "AS15169",
			ProbeCC:         "US",
			SoftwareName:    "miniooni",
			SoftwareVersion: "0.1.0-dev",
			SupportedTests: []string{
				"web_connectivity",
			},
		},
		Password:  "xx",
		UserAgent: "miniooni/0.1.0-dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("result should not be nil here")
	}
	if result.ClientID == "" {
		t.Fatal("ClientID should not be empty")
	}
}

func TestIntegrationFailure(t *testing.T) {
	// The successful integration test contains the minimal amount
	// of fields expected by the orchestra. Any less amount of fields,
	// such as we do here, results in the API returning error.
	result, err := Do(context.Background(), Config{
		BaseURL:    "https://ps-test.ooni.io",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if result != nil {
		t.Fatal("result should be nil here")
	}
}
