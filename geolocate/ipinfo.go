package geolocate

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

// IPInfoResponse is the response returned by ipinfo.io
type IPInfoResponse struct {
	IP string `json:"ip"`
}

// IPInfoIPLookup performs the IP lookup using ipinfo.io
func IPInfoIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	var v IPInfoResponse
	err := (httpx.Client{
		Accept:     "application/json",
		BaseURL:    "https://ipinfo.io",
		HTTPClient: httpClient,
		Logger:     logger,
		UserAgent:  httpheader.CLIUserAgent(), // we must be a CLI client
	}).GetJSON(ctx, "/", &v)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	return v.IP, nil
}
