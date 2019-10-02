// Package invalid returns an invalid IP.
package invalid

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/log"
)

// Do performs the IP lookup.
func Do(
	ctx context.Context,
	httpClient *http.Client,
	logger log.Logger,
	userAgent string,
) (string, error) {
	return "invalid IP", nil
}
