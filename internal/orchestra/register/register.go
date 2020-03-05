// Package register contains code to register to the OONI orchestra.
package register

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/model"
)

// Config contains configs for registering to OONI orchestra.
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     model.Logger
	Metadata   metadata.Metadata
	Password   string
	UserAgent  string
}

type request struct {
	metadata.Metadata
	Password string `json:"password"`
}

// Result contains the result of logging in.
type Result struct {
	ClientID string `json:"client_id"`
}

// Do registers this probe with OONI orchestra
func Do(ctx context.Context, config Config) (*Result, error) {
	req := &request{
		Metadata: config.Metadata,
		Password: config.Password,
	}
	var resp Result
	err := (&jsonapi.Client{
		BaseURL:    config.BaseURL,
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		UserAgent:  config.UserAgent,
	}).Create(ctx, "/api/v1/register", req, &resp)
	if err != nil {
		return nil, err
	}
	// Implementation note: the API does not return 200 unless there
	// is success, so we don't bother with reading the error field
	return &resp, nil
}
