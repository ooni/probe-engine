package login

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/orchestra/register"
)

const password = "xx"

func TestIntegrationSuccess(t *testing.T) {
	clientID, err := doRegister()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Do(context.Background(), Config{
		BaseURL:    "https://ps-test.ooni.io",
		ClientID:   clientID,
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		Password:   password,
		UserAgent:  "miniooni/0.1.0-dev",
	})
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

func TestIntegrationFailure(t *testing.T) {
	// This should fail because the username/password is wrong
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

func doRegister() (string, error) {
	result, err := register.Do(context.Background(), register.Config{
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
		Password:  password,
		UserAgent: "miniooni/0.1.0-dev",
	})
	if err != nil {
		return "", err
	}
	return result.ClientID, nil
}
