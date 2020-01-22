// Package tor contains code to fetch targets for the tor experiment.
package tor

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ooni/probe-engine/httpx/jsonapi"
	"github.com/ooni/probe-engine/internal/orchestra/login"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
)

// Config contains settings.
type Config struct {
	Auth       *login.Auth
	BaseURL    string
	HTTPClient *http.Client
	Logger     log.Logger
	UserAgent  string
}

// Query retrieves the tor experiment targets. This function will either
// return a nonzero list of targets or an error.
func Query(ctx context.Context, config Config) (result map[string]model.TorTarget, err error) {
	if config.Auth == nil {
		return nil, errors.New("config.Auth is nil")
	}
	authorization := fmt.Sprintf("Bearer %s", config.Auth.Token)
	err = (&jsonapi.Client{
		Authorization: authorization,
		BaseURL:       config.BaseURL,
		HTTPClient:    config.HTTPClient,
		Logger:        config.Logger,
		UserAgent:     config.UserAgent,
	}).Read(ctx, "/api/v1/test-list/tor-targets", &result)
	return
}
