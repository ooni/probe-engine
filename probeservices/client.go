package probeservices

import (
	"errors"
	"net/url"

	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/model"
)

// Client is a client for the OONI probe services API.
type Client struct {
	jsonapi.Client
}

var (
	// ErrUnsupportedEndpoint indicates that we don't support this endpoint type.
	ErrUnsupportedEndpoint = errors.New("probe services: unsupported endpoint type")

	// ErrUnsupportedCloudFrontAddress indicates that we don't support this
	// cloudfront address (e.g. wrong scheme, presence of port).
	ErrUnsupportedCloudFrontAddress = errors.New(
		"probe services: unsupported cloud front address",
	)
)

// NewClient creates a new client for the specified probe services endpoint.
func NewClient(sess model.ExperimentSession, endpoint model.Service) (*Client, error) {
	client := &Client{Client: jsonapi.Client{
		BaseURL:    endpoint.Address,
		HTTPClient: sess.DefaultHTTPClient(),
		Logger:     sess.Logger(),
		ProxyURL:   sess.ProxyURL(),
		UserAgent:  sess.UserAgent(),
	}}
	switch endpoint.Type {
	case "https":
		return client, nil
	case "cloudfront":
		// Do the cloudfronting dance. The front must appear inside of the
		// URL, so that we use it for DNS resolution and SNI. The real domain
		// must instead appear inside of the Host header.
		URL, err := url.Parse(client.BaseURL)
		if err != nil {
			return nil, err
		}
		if URL.Scheme != "https" || URL.Host != URL.Hostname() {
			return nil, ErrUnsupportedCloudFrontAddress
		}
		client.Client.Host = URL.Hostname()
		URL.Host = endpoint.Front
		client.BaseURL = URL.String()
		if _, err := url.Parse(client.BaseURL); err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, ErrUnsupportedEndpoint
	}
}

// Default returns the default probe services
func Default() []model.Service {
	return []model.Service{{
		Address: "https://bouncer.ooni.io",
		Type:    "https",
	}, {
		Front:   "dkyhjv0wpi2dk.cloudfront.net",
		Type:    "cloudfront",
		Address: "https://dkyhjv0wpi2dk.cloudfront.net",
	}}
}
