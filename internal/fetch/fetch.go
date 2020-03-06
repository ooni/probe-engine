// Package fetch is used to fetch resources.
package fetch

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ooni/probe-engine/model"
)

// Client is a client for fetching resources.
type Client struct {
	// Authorization is the authorization header to use.
	Authorization string

	// HTTPClient is the http client to use.
	HTTPClient *http.Client

	// Logger is the logger to use.
	Logger model.Logger

	// UserAgent is the user agent to use.
	UserAgent string
}

func (c *Client) makeRequest(
	ctx context.Context, URL string,
) (*http.Request, error) {
	c.Logger.Debugf("fetch: URL: %s", URL)
	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", c.UserAgent)
	if c.Authorization != "" {
		request.Header.Set("Authorization", c.Authorization)
	}
	return request.WithContext(ctx), nil
}

func (c *Client) do(request *http.Request) ([]byte, error) {
	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("Request failed: %s", response.Status)
	}
	return ioutil.ReadAll(response.Body)
}

// Fetch fetches the specified resource and returns it.
func (c *Client) Fetch(
	ctx context.Context, URL string,
) ([]byte, error) {
	request, err := c.makeRequest(ctx, URL)
	if err != nil {
		return nil, err
	}
	return c.do(request)
}

// FetchAndVerify fetches and verifies a specific resource.
func (c *Client) FetchAndVerify(
	ctx context.Context, URL, SHA256Sum string,
) ([]byte, error) {
	c.Logger.Debugf("fetch: expected SHA256: %s", SHA256Sum)
	data, err := c.Fetch(ctx, URL)
	if err != nil {
		return nil, err
	}
	s := fmt.Sprintf("%x", sha256.Sum256(data))
	c.Logger.Debugf("fetch: real SHA256: %s", s)
	if SHA256Sum != s {
		return nil, fmt.Errorf(
			"SHA256 mismatch: got %s and expected %s", s, SHA256Sum,
		)
	}
	return data, nil
}
