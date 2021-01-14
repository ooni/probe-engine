package geolocate

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/model"
)

func invalidIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	return "invalid IP", nil
}
