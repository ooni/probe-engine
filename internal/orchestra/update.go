package orchestra

import (
	"context"
	"fmt"

	"github.com/ooni/probe-engine/internal/httpx"
)

type updateRequest struct {
	Metadata
}

// Update updates the state of a probe
func (c Client) Update(ctx context.Context, metadata Metadata) error {
	if !metadata.Valid() {
		return errInvalidMetadata
	}
	creds, auth, err := c.getCredsAndAuth()
	if err != nil {
		return err
	}
	authorization := fmt.Sprintf("Bearer %s", auth.Token)
	req := &updateRequest{Metadata: metadata}
	var resp struct{}
	urlpath := fmt.Sprintf("/api/v1/update/%s", creds.ClientID)
	return (httpx.Client{
		Authorization: authorization,
		BaseURL:       c.BaseURL,
		HTTPClient:    c.HTTPClient,
		Logger:        c.Logger,
		UserAgent:     c.UserAgent,
	}).UpdateJSON(ctx, urlpath, req, &resp)
}
