package probeservices

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

var (
	// ErrUnsupportedEndpoint indicates that we don't support this endpoint type.
	ErrUnsupportedEndpoint = errors.New("probe services: unsupported endpoint type")

	// ErrUnsupportedCloudFrontAddress indicates that we don't support this
	// cloudfront address (e.g. wrong scheme, presence of port).
	ErrUnsupportedCloudFrontAddress = errors.New(
		"probe services: unsupported cloud front address",
	)

	errInvalidMetadata = errors.New("invalid metadata")
	errNotLoggedIn     = errors.New("not logged in")
	errNotRegistered   = errors.New("not registered")
)

// Client is a client for the OONI probe services API.
type Client struct {
	httpx.Client
	LoginCalls    *atomicx.Int64
	RegisterCalls *atomicx.Int64
	StateFile     *StateFile
}

func (c Client) getCredsAndAuth() (*LoginCredentials, *LoginAuth, error) {
	state := c.StateFile.Get()
	creds := state.Credentials()
	if creds == nil {
		return nil, nil, errNotRegistered
	}
	auth := state.Auth()
	if auth == nil {
		return nil, nil, errNotLoggedIn
	}
	return creds, auth, nil
}

// NewClient creates a new client for the specified probe services endpoint.
func NewClient(sess model.ExperimentSession, endpoint model.Service) (*Client, error) {
	client := &Client{
		Client: httpx.Client{
			BaseURL:    endpoint.Address,
			HTTPClient: sess.DefaultHTTPClient(),
			Logger:     sess.Logger(),
			ProxyURL:   sess.ProxyURL(),
			UserAgent:  sess.UserAgent(),
		},
		LoginCalls:    atomicx.NewInt64(),
		RegisterCalls: atomicx.NewInt64(),
		StateFile:     NewStateFile(sess.KeyValueStore()),
	}
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
		Address: "https://ps.ooni.io",
		Type:    "https",
	}, {
		Front:   "dkyhjv0wpi2dk.cloudfront.net",
		Type:    "cloudfront",
		Address: "https://dkyhjv0wpi2dk.cloudfront.net",
	}}
}

// SortEndpoints gives priority to https, then cloudfronted, then onion.
func SortEndpoints(in []model.Service) (out []model.Service) {
	for _, entry := range in {
		if entry.Type == "https" {
			out = append(out, entry)
		}
	}
	for _, entry := range in {
		if entry.Type == "cloudfront" {
			out = append(out, entry)
		}
	}
	for _, entry := range in {
		if entry.Type == "onion" {
			out = append(out, entry)
		}
	}
	return
}

// OnlyHTTPS returns the HTTPS endpoints only.
func OnlyHTTPS(in []model.Service) (out []model.Service) {
	for _, entry := range in {
		if entry.Type == "https" {
			out = append(out, entry)
		}
	}
	return
}

// OnlyFallbacks returns the fallback endpoints only.
func OnlyFallbacks(in []model.Service) (out []model.Service) {
	for _, entry := range SortEndpoints(in) {
		if entry.Type != "https" {
			out = append(out, entry)
		}
	}
	return
}

// Candidate is a candidate probe service.
type Candidate struct {
	// Duration is the time it took to access the service.
	Duration time.Duration

	// Err indicates whether the service works.
	Err error

	// Endpoint is the service endpoint.
	Endpoint model.Service

	// TestHelpers contains the data returned by the endpoint.
	TestHelpers map[string][]model.Service
}

func (c *Candidate) try(ctx context.Context, sess model.ExperimentSession) {
	client, err := NewClient(sess, c.Endpoint)
	if err != nil {
		c.Err = err
		return
	}
	start := time.Now()
	testhelpers, err := client.GetTestHelpers(ctx)
	c.Duration = time.Now().Sub(start)
	c.Err = err
	c.TestHelpers = testhelpers
	sess.Logger().Debugf("probe services: %+v: %+v %s", c.Endpoint, err, c.Duration)
}

func try(ctx context.Context, sess model.ExperimentSession, svc model.Service) *Candidate {
	candidate := &Candidate{Endpoint: svc}
	candidate.try(ctx, sess)
	return candidate
}

// TryAll tries all the input services using the provided context and session. It
// returns a list containing information on each candidate that was tried. We will
// try all the HTTPS candidates first. So, the beginning of the list will contain
// all of them, and for each of them you will know whether it worked (by checking the
// Err field) and how fast it was (by checking the Duration field). You should pick
// the fastest one that worked. If none of them works, then TryAll will subsequently
// attempt with all the available fallbacks, and return at the first success. In
// such case, you will see a list of N failing HTTPS candidates, followed by a single
// successful fallback candidate (e.g. cloudfronted). If all candidates fail, you
// see in output a list containing all entries where Err is not nil.
func TryAll(ctx context.Context, sess model.ExperimentSession, in []model.Service) (out []*Candidate) {
	var found bool
	for _, svc := range OnlyHTTPS(in) {
		candidate := try(ctx, sess, svc)
		out = append(out, candidate)
		if candidate.Err == nil {
			found = true
		}
	}
	if !found {
		for _, svc := range OnlyFallbacks(in) {
			candidate := try(ctx, sess, svc)
			out = append(out, candidate)
			if candidate.Err == nil {
				return
			}
		}
	}
	return
}

// SelectBest selects the best among the candidates. If there is no
// suitable candidate, then this function returns nil.
func SelectBest(candidates []*Candidate) (selected *Candidate) {
	for _, e := range candidates {
		if e.Err != nil {
			continue
		}
		if selected == nil {
			selected = e
			continue
		}
		if selected.Duration > e.Duration {
			selected = e
			continue
		}
	}
	return
}
