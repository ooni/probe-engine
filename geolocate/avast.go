// Package geolocate implements IP lookup, resolver lookup, and GeoIP
// location an OONI Probe instance.
package geolocate

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

// AvastResponse is the response returned by Avast IP lookup services.
type AvastResponse struct {
	IP string `json:"ip"`
}

// AvastIPLookup performs the IP lookup using Avast services.
func AvastIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	var v AvastResponse
	err := (httpx.Client{
		BaseURL:    "https://ip-info.ff.avast.com",
		HTTPClient: httpClient,
		Logger:     logger,
		UserAgent:  userAgent,
	}).GetJSON(ctx, "/v1/info", &v)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	return v.IP, nil
}
