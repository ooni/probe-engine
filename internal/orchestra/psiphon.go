package orchestra

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

// PsiphonConfig contains configs for fetching psiphon config
type PsiphonConfig struct {
	Auth       *login.Auth
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
