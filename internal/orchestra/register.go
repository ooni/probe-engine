package orchestra

import (
	"context"

	"github.com/ooni/probe-engine/internal/httpx"
)

type registerRequest struct {
	Metadata
	Password string `json:"password"`
}

type registerResult struct {
	ClientID string `json:"client_id"`
}

// MaybeRegister registers this client if not already registered
func (c Client) MaybeRegister(ctx context.Context, metadata Metadata) error {
	if !metadata.Valid() {
		return errInvalidMetadata
	}
	state := c.StateFile.Get()
	if state.Credentials() != nil {
		return nil // we're already good
	}
	c.RegisterCalls.Add(1)
	pwd := randomPassword(64)
	req := &registerRequest{
		Metadata: metadata,
		Password: pwd,
	}
	var resp registerResult
	err := (httpx.Client{
		BaseURL:    c.BaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	}).CreateJSON(ctx, "/api/v1/register", req, &resp)
	if err != nil {
		return err
	}
	state.ClientID = resp.ClientID
	state.Password = pwd
	return c.StateFile.Set(state)
}
