// Package orchestra contains common orchestra code. Orchestra is a set of
// OONI APIs for probe orchestration. You can find code implementing each
// specific API into a subpackage of this package. This package contains the
// toplevel orchestra client that the session should use.
package orchestra

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/internal/orchestra/login"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/orchestra/register"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/orchestra/testlists/psiphon"
	"github.com/ooni/probe-engine/internal/orchestra/update"
	"github.com/ooni/probe-engine/log"
)

// Client is a client for OONI orchestra
type Client struct {
	HTTPClient       *http.Client
	Logger           log.Logger
	OrchestraBaseURL string
	RegistryBaseURL  string
	StateFile        statefile.StateFile
	UserAgent        string
	registerCalls    int
	loginCalls       int
}

// NewClient creates a new client.
func NewClient(
	httpClient *http.Client, logger log.Logger,
	userAgent string, stateFile statefile.StateFile,
) *Client {
	return &Client{
		HTTPClient:       httpClient,
		Logger:           logger,
		OrchestraBaseURL: "https://orchestrate.ooni.io",
		RegistryBaseURL:  "https://registry.ooni.io",
		StateFile:        stateFile,
		UserAgent:        userAgent,
	}
}

var (
	errInvalidMetadata = errors.New("invalid metadata")
	errNotLoggedIn     = errors.New("not logged in")
	errNotRegistered   = errors.New("not registered")
)

// MaybeRegister registers this client if not already registered
func (c *Client) MaybeRegister(
	ctx context.Context, metadata metadata.Metadata,
) error {
	if !metadata.Valid() {
		return errInvalidMetadata
	}
	state := c.StateFile.Get()
	if state.Credentials() != nil {
		return nil // we're already good
	}
	c.registerCalls++
	pwd := randomPassword(64)
	result, err := register.Do(ctx, register.Config{
		BaseURL:    c.RegistryBaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		Metadata:   metadata,
		Password:   pwd,
		UserAgent:  c.UserAgent,
	})
	if err != nil {
		return err
	}
	state.ClientID = result.ClientID
	state.Password = pwd
	return c.StateFile.Set(state)
}

// MaybeLogin performs login if necessary
func (c *Client) MaybeLogin(ctx context.Context) error {
	state := c.StateFile.Get()
	if state.Auth() != nil {
		return nil // we're already good
	}
	creds := state.Credentials()
	if creds == nil {
		return errNotRegistered
	}
	c.loginCalls++
	auth, err := login.Do(ctx, login.Config{
		BaseURL:     c.RegistryBaseURL,
		Credentials: *creds,
		HTTPClient:  c.HTTPClient,
		Logger:      c.Logger,
		UserAgent:   c.UserAgent,
	})
	if err != nil {
		return err
	}
	state.Expire = auth.Expire
	state.Token = auth.Token
	return c.StateFile.Set(state)
}

func (c *Client) getCredsAndAuth() (*login.Credentials, *login.Auth, error) {
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

// Update updates the state of a probe
func (c *Client) Update(
	ctx context.Context, metadata metadata.Metadata,
) error {
	if !metadata.Valid() {
		return errInvalidMetadata
	}
	creds, auth, err := c.getCredsAndAuth()
	if err != nil {
		return err
	}
	return update.Do(context.Background(), update.Config{
		Auth:       auth,
		BaseURL:    c.OrchestraBaseURL,
		ClientID:   creds.ClientID,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		Metadata:   metadata,
		UserAgent:  c.UserAgent,
	})
}

// FetchPsiphonConfig fetches psiphon config from authenticated OONI orchestra.
func (c *Client) FetchPsiphonConfig(ctx context.Context) ([]byte, error) {
	_, auth, err := c.getCredsAndAuth()
	if err != nil {
		return nil, err
	}
	return psiphon.Query(ctx, psiphon.Config{
		Auth:       auth,
		BaseURL:    c.OrchestraBaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	})
}

func randomPassword(n int) string {
	// See https://stackoverflow.com/questions/22892120
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rnd.Intn(len(letterBytes))]
	}
	return string(b)
}
