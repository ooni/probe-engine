// Package psiphon implements fetching psiphon config using orchestra
package psiphon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ooni/probe-engine/internal/fetch"
	"github.com/ooni/probe-engine/internal/orchestra/login"
	"github.com/ooni/probe-engine/internal/urlpath"
	"github.com/ooni/probe-engine/model"
)

// Config contains configs for fetching psiphon config
type Config struct {
	Auth       *login.Auth
	BaseURL    string
	HTTPClient *http.Client
	Logger     model.Logger
	UserAgent  string
}

// Query retrieves the psiphon config
func Query(ctx context.Context, config Config) ([]byte, error) {
	if config.Auth == nil {
		return nil, errors.New("config.Auth is nil")
	}
	url, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, err
	}
	url.Path = urlpath.Append(url.Path, "/api/v1/test-list/psiphon-config")
	authorization := fmt.Sprintf("Bearer %s", config.Auth.Token)
	return (&fetch.Client{
		Authorization: authorization,
		HTTPClient:    config.HTTPClient,
		Logger:        config.Logger,
		UserAgent:     config.UserAgent,
	}).Fetch(ctx, url.String())
}
