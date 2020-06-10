package orchestra_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/testorchestra"
)

func TestLoginSuccess(t *testing.T) {
	clientID, err := testorchestra.Register()
	if err != nil {
		t.Fatal(err)
	}
	result, err := testorchestra.Login(clientID)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("result should not be nil here")
	}
	if result.Expire.IsZero() {
		t.Fatal("Expire should not be zero")
	}
	if result.Token == "" {
		t.Fatal("Token should not be empty")
	}
}

func TestLoginFailure(t *testing.T) {
	// This should fail because the username/password is wrong
	result, err := orchestra.Login(context.Background(), orchestra.LoginConfig{
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
