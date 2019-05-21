// Package ubuntu lookups the IP using Ubuntu.
package ubuntu

import (
	"context"
	"encoding/xml"
	"net/http"

	"github.com/ooni/probe-engine/httpx/fetch"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
)

type response struct {
	XMLName xml.Name `xml:"Response"`
	IP      string   `xml:"Ip"`
}

// Do performs the IP lookup.
func Do(
	ctx context.Context,
	httpClient *http.Client,
	logger log.Logger,
	userAgent string,
) (string, error) {
	data, err := (&fetch.Client{
		HTTPClient: httpClient,
		Logger:     logger,
		UserAgent:  userAgent,
	}).Fetch(ctx, "https://geoip.ubuntu.com/lookup")
	if err != nil {
		return model.DefaultProbeIP, err
	}
	logger.Debugf("ubuntu: body: %s", string(data))
	var v response
	err = xml.Unmarshal(data, &v)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	return v.IP, nil
}
