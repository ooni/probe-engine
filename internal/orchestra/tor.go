package orchestra

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

// TorConfig contains settings.
type TorConfig struct {
	Auth       *LoginAuth
	BaseURL    string
	HTTPClient *http.Client
	Logger     model.Logger
	UserAgent  string
}

// TorQuery retrieves the tor experiment targets.
func TorQuery(ctx context.Context, config TorConfig) (result map[string]model.TorTarget, err error) {
	if config.Auth == nil {
		return nil, errors.New("config.Auth is nil")
	}
	authorization := fmt.Sprintf("Bearer %s", config.Auth.Token)
	err = (&httpx.Client{
		Authorization: authorization,
		BaseURL:       config.BaseURL,
		HTTPClient:    config.HTTPClient,
		Logger:        config.Logger,
		UserAgent:     config.UserAgent,
	}).ReadJSON(ctx, "/api/v1/test-list/tor-targets", &result)
	return
}
