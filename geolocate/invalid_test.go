package geolocate

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/model"
)

// InvalidIPLookup is an IP lookup that always returns an invalid IP.
func InvalidIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	return "invalid IP", nil
}
