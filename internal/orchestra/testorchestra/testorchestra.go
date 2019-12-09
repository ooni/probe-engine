// Package testorchestra contains code to simplify testing
package testorchestra

import (
	"context"
	"net/http"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra/login"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/orchestra/register"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/orchestra/update"
)

const password = "xx"

// Register register a fictional probe and returns the clientID
// on success and an error on failure.
func Register() (string, error) {
	result, err := register.Do(context.Background(), register.Config{
		BaseURL:    "https://ps-test.ooni.io",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		Metadata:   MetadataFixture(),
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
func Login(clientID string) (*login.Auth, error) {
	return login.Do(context.Background(), login.Config{
		BaseURL: "https://ps-test.ooni.io",
		Credentials: login.Credentials{
			ClientID: clientID,
			Password: password,
		},
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "miniooni/0.1.0-dev",
	})
}

// Update updates information about a probe
func Update(auth *login.Auth, clientID string) error {
	return update.Do(context.Background(), update.Config{
		Auth:       auth,
		BaseURL:    "https://ps-test.ooni.io",
		ClientID:   clientID,
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		Metadata:   MetadataFixture(),
		UserAgent:  "miniooni/0.1.0-dev",
	})
}

// MetadataFixture returns a valid metadata struct
func MetadataFixture() metadata.Metadata {
	return metadata.Metadata{
		Platform:        "linux",
		ProbeASN:        "AS15169",
		ProbeCC:         "US",
		SoftwareName:    "miniooni",
		SoftwareVersion: "0.1.0-dev",
		SupportedTests: []string{
			"web_connectivity",
		},
	}
}

// StateFileFake is a fake state file
type StateFileFake struct {
	GetState *statefile.State
	GetError error
	SetError error
}

// Set overrides the current state
func (sf *StateFileFake) Set(s *statefile.State) error {
	return nil
}

// Get returns the current state
func (sf *StateFileFake) Get() (*statefile.State, error) {
	return sf.GetState, sf.GetError
}
