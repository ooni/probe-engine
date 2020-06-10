package orchestra

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/internal/orchestra/login"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/model"
)

// UpdateConfig contains configs for calling the Update API.
type UpdateConfig struct {
	Auth       *login.Auth
	BaseURL    string
	ClientID   string
	HTTPClient *http.Client
	Logger     model.Logger
	Metadata   metadata.Metadata
	UserAgent  string
}

type request struct {
	metadata.Metadata
}

// Update updates OONI orchestra view of this probe
func Update(ctx context.Context, config UpdateConfig) error {
	if config.Auth == nil {
		return errors.New("config.Auth is nil")
	}
	authorization := fmt.Sprintf("Bearer %s", config.Auth.Token)
	req := &request{Metadata: config.Metadata}
	var resp struct{}
	urlpath := fmt.Sprintf("/api/v1/update/%s", config.ClientID)
	return (&jsonapi.Client{
		Authorization: authorization,
		BaseURL:       config.BaseURL,
		HTTPClient:    config.HTTPClient,
		Logger:        config.Logger,
		UserAgent:     config.UserAgent,
	}).Update(ctx, urlpath, req, &resp)
}
