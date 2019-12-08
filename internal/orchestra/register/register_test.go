package register_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra/register"
	"github.com/ooni/probe-engine/internal/orchestra/testorchestra"
)

func TestIntegrationSuccess(t *testing.T) {
	clientID, err := testorchestra.Register()
	if err != nil {
		t.Fatal(err)
	}
	if clientID == "" {
		t.Fatal("ClientID should not be empty")
	}
}

func TestIntegrationFailure(t *testing.T) {
	// The successful integration test contains the minimal amount
	// of fields expected by the orchestra. Any less amount of fields,
	// such as we do here, results in the API returning error.
	result, err := register.Do(context.Background(), register.Config{
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
