package orchestra

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

// PsiphonConfig contains configs for fetching psiphon config
type PsiphonConfig struct {
	Auth       *LoginAuth
	BaseURL    string
	HTTPClient *http.Client
	Logger     model.Logger
	UserAgent  string
}

// PsiphonQuery retrieves the psiphon config
func PsiphonQuery(ctx context.Context, config PsiphonConfig) ([]byte, error) {
	if config.Auth == nil {
		return nil, errors.New("config.Auth is nil")
	}
	authorization := fmt.Sprintf("Bearer %s", config.Auth.Token)
	return (&httpx.Client{
		Authorization: authorization,
		BaseURL:       config.BaseURL,
		HTTPClient:    config.HTTPClient,
		Logger:        config.Logger,
		UserAgent:     config.UserAgent,
	}).FetchResource(ctx, "/api/v1/test-list/psiphon-config")
}
