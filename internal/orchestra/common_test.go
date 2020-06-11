package orchestra_test

import (
	"context"
	"net/http"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/orchestra"
)

const password = "xx"

// Register registers a fictional probe and returns the clientID
// on success and an error on failure.
func Register() (string, error) {
	result, err := orchestra.Register(context.Background(), orchestra.RegisterConfig{
		BaseURL:    "https://ps-test.ooni.io",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		Metadata:   mockable.OrchestraMetadataFixture(),
		Password:   password,
		UserAgent:  "miniooni/0.1.0-dev",
	})
	if err != nil {
		return "", err
	}
	return result.ClientID, nil
}

// Login performs a login and returns the authentication token
// information on success, and an error on failure.
func Login(clientID string) (*orchestra.LoginAuth, error) {
	return orchestra.Login(context.Background(), orchestra.LoginConfig{
		BaseURL: "https://ps-test.ooni.io",
		Credentials: orchestra.LoginCredentials{
			ClientID: clientID,
			Password: password,
		},
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	})
}

// Update updates information about a probe
func Update(auth *orchestra.LoginAuth, clientID string) error {
	return orchestra.Update(context.Background(), orchestra.UpdateConfig{
		Auth:       auth,
		BaseURL:    "https://ps-test.ooni.io",
		ClientID:   clientID,
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		Metadata:   mockable.OrchestraMetadataFixture(),
		UserAgent:  "miniooni/0.1.0-dev",
	})
}
