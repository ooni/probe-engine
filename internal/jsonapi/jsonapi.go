// Package jsonapi interacts with HTTP JSON APIs. In OONI we use
// this code when accessing API like, e.g., the OONI collector.
package jsonapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/dialer"
)

// Client is a client for a JSON API.
type Client struct {
	// Authorization contains the authorization header.
	Authorization string

	// BaseURL is the base URL of the API.
	BaseURL string

	// HTTPClient is the http client to use.
	HTTPClient *http.Client

	// Host allows to set a specific host header. This is useful
	// to implement, e.g., cloudfronting.
	Host string

	// Logger is the logger to use.
	Logger model.Logger

	// ProxyURL allows to force a proxy URL to fallback to a
	// tunnel, e.g., Psiphon.
	ProxyURL *url.URL

	// UserAgent is the user agent to use.
	UserAgent string
}

// NewRequest creates a new request with a JSON body
func (c Client) NewRequest(
	ctx context.Context, method, resourcePath string,
	query url.Values, body interface{}) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	c.Logger.Debugf("jsonapi: request body: %d bytes", len(data))
	return c.newRequestWithSerializedJSONBody(
		ctx, method, resourcePath, query, bytes.NewReader(data))
}

func (c Client) newRequestWithSerializedJSONBody(
	ctx context.Context, method, resourcePath string,
	query url.Values, body io.Reader) (*http.Request, error) {
	URL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	URL.Path = resourcePath
	if query != nil {
		URL.RawQuery = query.Encode()
	}
	c.Logger.Debugf("jsonapi: method: %s", method)
	c.Logger.Debugf("jsonapi: URL: %s", URL.String())
	request, err := http.NewRequest(method, URL.String(), body)
	if err != nil {
		return nil, err
	}
	request.Host = c.Host // allow cloudfronting
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.Authorization != "" {
		request.Header.Set("Authorization", c.Authorization)
	}
	request.Header.Set("User-Agent", c.UserAgent)
	// Implementation note: the following allows tunneling if c.ProxyURL
	// is not nil. Because the proxy URL is set as part of each request
	// generated using this function, every request that eventually needs
	// to reconnect will always do so using the proxy.
	ctx = dialer.WithProxyURL(ctx, c.ProxyURL)
	return request.WithContext(ctx), nil
}

// Do performs the provided request and unmarshals the JSON response body
// into the provided output variable.
func (c Client) Do(request *http.Request, output interface{}) error {
	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return fmt.Errorf("jsonapi: request failed: %s", response.Status)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	c.Logger.Debugf("jsonapi: response body: %d bytes", len(data))
	return json.Unmarshal(data, output)
}

// Read reads the JSON resource at resourcePath and unmarshals the
// results into output. The request is bounded by the lifetime of the
// context passed as argument. Returns the error that occurred.
func (c Client) Read(ctx context.Context, resourcePath string, output interface{}) error {
	return c.ReadWithQuery(ctx, resourcePath, nil, output)
}

// ReadWithQuery is like Read but also has a query.
func (c Client) ReadWithQuery(
	ctx context.Context, resourcePath string,
	query url.Values, output interface{}) error {
	request, err := c.newRequestWithSerializedJSONBody(ctx, "GET", resourcePath, query, nil)
	if err != nil {
		return err
	}
	return c.Do(request, output)
}

// Create creates a JSON subresource of the resource at resourcePath
// using the JSON document at input and returning the result into the
// JSON document at output. The request is bounded by the context's
// lifetime. Returns the error that occurred.
func (c Client) Create(
	ctx context.Context, resourcePath string, input, output interface{}) error {
	request, err := c.NewRequest(ctx, "POST", resourcePath, nil, input)
	if err != nil {
		return err
	}
	return c.Do(request, output)
}

// Update updates a JSON resource at a specific path and returns
// the error that occurred and possibly an output document
func (c Client) Update(
	ctx context.Context, resourcePath string, input, output interface{}) error {
	request, err := c.NewRequest(ctx, "PUT", resourcePath, nil, input)
	if err != nil {
		return err
	}
	return c.Do(request, output)
}
