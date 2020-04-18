// Package mlablocate contains a locate.measurementlab.net client.
package mlablocate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ooni/probe-engine/model"
)

// Client is a locate.measurementlab.net client.
type Client struct {
	HTTPClient *http.Client
	Hostname   string
	Logger     model.Logger
	Scheme     string
	UserAgent  string
}

// NewClient creates a new locate.measurementlab.net client.
func NewClient(httpClient *http.Client, logger model.Logger, userAgent string) *Client {
	return &Client{
		HTTPClient: httpClient,
		Hostname:   "locate.measurementlab.net",
		Logger:     logger,
		Scheme:     "https",
		UserAgent:  userAgent,
	}
}

type locateResult struct {
	FQDN string `json:"fqdn"`
}

// Query performs a locate.measurementlab.net query.
func (c *Client) Query(ctx context.Context, tool string) (string, error) {
	URL := &url.URL{
		Scheme: c.Scheme,
		Host:   c.Hostname,
		Path:   tool,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", URL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", c.UserAgent)
	c.Logger.Debugf("mlablocate: GET %s", URL.String())
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("mlablocate: non-200 status code: %d", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf("mlablocate: %s", string(data))
	var result locateResult
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if result.FQDN == "" {
		return "", errors.New("mlablocate: returned empty FQDN")
	}
	return result.FQDN, nil
}
