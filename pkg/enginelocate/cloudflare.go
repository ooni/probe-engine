package enginelocate

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/ooni/probe-engine/pkg/httpx"
	"github.com/ooni/probe-engine/pkg/model"
)

func cloudflareIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
	resolver model.Resolver,
) (string, error) {
	data, err := (&httpx.APIClientTemplate{
		BaseURL:    "https://www.cloudflare.com",
		HTTPClient: httpClient,
		Logger:     logger,
		UserAgent:  model.HTTPHeaderUserAgent,
	}).WithBodyLogging().Build().FetchResource(ctx, "/cdn-cgi/trace")
	if err != nil {
		return model.DefaultProbeIP, err
	}
	r := regexp.MustCompile("(?:ip)=(.*)")
	ip := strings.Trim(string(r.Find(data)), "ip=")
	logger.Debugf("cloudflare: body: %s", ip)
	return ip, nil
}
