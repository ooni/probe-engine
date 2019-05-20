// Package avast lookups the IP using avast.
package avast

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/geoiplookup/constants"
	"github.com/ooni/probe-engine/httpx/jsonapi"
	"github.com/ooni/probe-engine/log"
)

type response struct {
	IP string `json:"ip"`
}

// Do performs the IP lookup.
func Do(
	ctx context.Context,
	httpClient *http.Client,
	logger log.Logger,
	userAgent string,
) (string, error) {
	var v response
	err := (&jsonapi.Client{
		BaseURL:    "https://ip-info.ff.avast.com",
		HTTPClient: httpClient,
		Logger:     logger,
		UserAgent:  userAgent,
	}).Read(ctx, "/v1/info", &v)
	if err != nil {
		return constants.DefaultProbeIP, err
	}
	return v.IP, nil
}
