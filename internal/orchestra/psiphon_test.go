package orchestra_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra/login"
	"github.com/ooni/probe-engine/internal/orchestra"
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
	data, err := orchestra.PsiphonQuery(context.Background(), orchestra.PsiphonConfig{
		Auth:       auth,
		BaseURL:    "https://ps-test.ooni.io",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	var config interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
}

func TestUnitAuthNil(t *testing.T) {
	data, err := orchestra.PsiphonQuery(context.Background(), orchestra.PsiphonConfig{
		Auth:       nil,
		BaseURL:    "https://ps-test.ooni.io",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected no data here")
	}
}

func TestUnitConfigInvalidURL(t *testing.T) {
	orchestrateURL := "\t\t\t"
	data, err := orchestra.PsiphonQuery(context.Background(), orchestra.PsiphonConfig{
		Auth:       new(login.Auth),
		BaseURL:    orchestrateURL,
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected no data here")
	}
}
