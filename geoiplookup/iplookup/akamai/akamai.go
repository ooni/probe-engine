// Package akamai lookups the IP using akamai.
package akamai

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
)

type response struct {
	IP string `json:"ip"`
}

// Do performs the IP lookup.
func Do(
	ctx context.Context,
	httpClient *http.Client,
	logger log.Logger,
	userAgent string,
) (string, error) {
	req, err := http.NewRequest("GET", "https://a248.e.akamai.net/", nil)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	req.Host = "whatismyip.akamai.com" // domain fronted request
	req.Header.Set("User-Agent", userAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}
