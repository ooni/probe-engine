package tor_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra/testlists/tor"
	"github.com/ooni/probe-engine/internal/orchestra/testorchestra"
)

func TestIntegrationSuccess(t *testing.T) {
	clientID, err := testorchestra.Register()
	if err != nil {
		t.Fatal(err)
	}
	auth, err := testorchestra.Login(clientID)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := tor.Query(context.Background(), tor.Config{
		Auth:       auth,
		BaseURL:    "https://ps-test.ooni.io",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	if targets == nil {
		t.Fatal("expected non-nil targets here")
	}
}

func TestUnitAuthNil(t *testing.T) {
	targets, err := tor.Query(context.Background(), tor.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if targets != nil {
		t.Fatal("expected no targets here")
	}
}
