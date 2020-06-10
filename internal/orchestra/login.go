package orchestra

import (
	"context"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/model"
)

// LoginConfig contains configs for logging in with the OONI orchestra.
type LoginConfig struct {
	BaseURL     string
	Credentials LoginCredentials
	HTTPClient  *http.Client
	Logger      model.Logger
	UserAgent   string
}

// LoginCredentials contains the login credentials
type LoginCredentials struct {
	ClientID string `json:"username"`
	Password string `json:"password"`
}

// LoginAuth contains authentication info
type LoginAuth struct {
	Expire time.Time `json:"expire"`
	Token  string    `json:"token"`
}

// Login logs this probe in with OONI orchestra
func Login(ctx context.Context, config LoginConfig) (*LoginAuth, error) {
	var resp LoginAuth
	err := (&jsonapi.Client{
		BaseURL:    config.BaseURL,
		HTTPClient: config.HTTPClient,
		Logger:     config.Logger,
		UserAgent:  config.UserAgent,
	}).Create(ctx, "/api/v1/login", config.Credentials, &resp)
	if err != nil {
		return nil, err
	}
	// Implementation note: the API does not return 200 unless there
	// is success, so we don't bother with reading the error field
	return &resp, nil
}
