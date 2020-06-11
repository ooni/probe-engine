package orchestra_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra"
)

func TestTorSuccess(t *testing.T) {
	clientID, err := Register()
	if err != nil {
		t.Fatal(err)
	}
	auth, err := Login(clientID)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := orchestra.TorQuery(context.Background(), orchestra.TorConfig{
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

func TestTorAuthNil(t *testing.T) {
	targets, err := orchestra.TorQuery(context.Background(), orchestra.TorConfig{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if targets != nil {
		t.Fatal("expected no targets here")
	}
}
