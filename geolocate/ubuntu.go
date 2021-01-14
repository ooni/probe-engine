package geolocate

import (
	"context"
	"encoding/xml"
	"net/http"

	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

type ubuntuResponse struct {
	XMLName xml.Name `xml:"Response"`
	IP      string   `xml:"Ip"`
}

func ubuntuIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	data, err := (httpx.Client{
		BaseURL:    "https://geoip.ubuntu.com/",
		HTTPClient: httpClient,
		Logger:     logger,
		UserAgent:  userAgent,
	}).FetchResource(ctx, "/lookup")
	if err != nil {
		return model.DefaultProbeIP, err
	}
	logger.Debugf("ubuntu: body: %s", string(data))
	var v ubuntuResponse
	err = xml.Unmarshal(data, &v)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	return v.IP, nil
}
