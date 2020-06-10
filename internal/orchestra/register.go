package orchestra

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/model"
)

// RegisterConfig contains configs for registering to OONI orchestra.
type RegisterConfig struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     model.Logger
	Metadata   Metadata
	Password   string
	UserAgent  string
}

type registerRequest struct {
	Metadata
	Password string `json:"password"`
}

// RegisterResult contains the result of logging in.
type RegisterResult struct {
	ClientID string `json:"client_id"`
}

// Register registers this probe with OONI orchestra
func Register(ctx context.Context, config RegisterConfig) (*RegisterResult, error) {
	req := &registerRequest{
		Metadata: config.Metadata,
		Password: config.Password,
	}
	var resp RegisterResult
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
