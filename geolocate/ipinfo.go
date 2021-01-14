package geolocate

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

type ipInfoResponse struct {
	IP string `json:"ip"`
}

func ipInfoIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	var v ipInfoResponse
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
