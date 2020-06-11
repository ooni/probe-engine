package orchestra

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

// UpdateConfig contains configs for calling the Update API.
type UpdateConfig struct {
	Auth       *LoginAuth
	BaseURL    string
	ClientID   string
	HTTPClient *http.Client
	Logger     model.Logger
	Metadata   Metadata
	UserAgent  string
}

type updateRequest struct {
	Metadata
}

// Update updates OONI orchestra view of this probe
func Update(ctx context.Context, config UpdateConfig) error {
	if config.Auth == nil {
		return errors.New("config.Auth is nil")
	}
	authorization := fmt.Sprintf("Bearer %s", config.Auth.Token)
	req := &updateRequest{Metadata: config.Metadata}
	var resp struct{}
	urlpath := fmt.Sprintf("/api/v1/update/%s", config.ClientID)
	return (httpx.Client{
		Authorization: authorization,
		BaseURL:       config.BaseURL,
		HTTPClient:    config.HTTPClient,
		Logger:        config.Logger,
		UserAgent:     config.UserAgent,
	}).UpdateJSON(ctx, urlpath, req, &resp)
}
