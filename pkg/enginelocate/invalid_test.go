package enginelocate

import (
	"context"

	"github.com/ooni/probe-engine/pkg/model"
)

func invalidIPLookup(
	ctx context.Context,
	httpClient model.HTTPClient,
	logger model.Logger,
	userAgent string,
	resolver model.Resolver,
) (string, error) {
	return "invalid IP", nil
}
