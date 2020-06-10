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

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/model"
)

// Client is a client for OONI orchestra
type Client struct {
	HTTPClient         *http.Client
	Logger             model.Logger
	LoginCalls         *atomicx.Int64
	OrchestrateBaseURL string
	RegisterCalls      *atomicx.Int64
	RegistryBaseURL    string
	StateFile          *StateFile
	UserAgent          string
}

// NewClient creates a new client.
func NewClient(
	httpClient *http.Client, logger model.Logger,
	userAgent string, stateFile *StateFile,
) *Client {
	return &Client{
		HTTPClient:         httpClient,
		Logger:             logger,
		LoginCalls:         atomicx.NewInt64(),
		OrchestrateBaseURL: "https://ps.ooni.io",
		RegisterCalls:      atomicx.NewInt64(),
		RegistryBaseURL:    "https://ps.ooni.io",
		StateFile:          stateFile,
		UserAgent:          userAgent,
	}
}

var (
	errInvalidMetadata = errors.New("invalid metadata")
	errNotLoggedIn     = errors.New("not logged in")
	errNotRegistered   = errors.New("not registered")
)

// MaybeRegister registers this client if not already registered
func (c *Client) MaybeRegister(
	ctx context.Context, metadata Metadata,
) error {
	if !metadata.Valid() {
		return errInvalidMetadata
	}
	state := c.StateFile.Get()
	if state.Credentials() != nil {
		return nil // we're already good
	}
	c.RegisterCalls.Add(1)
	pwd := randomPassword(64)
	result, err := Register(ctx, RegisterConfig{
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
	c.LoginCalls.Add(1)
	auth, err := Login(ctx, LoginConfig{
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

func (c *Client) getCredsAndAuth() (*LoginCredentials, *LoginAuth, error) {
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
	ctx context.Context, metadata Metadata,
) error {
	if !metadata.Valid() {
		return errInvalidMetadata
	}
	creds, auth, err := c.getCredsAndAuth()
	if err != nil {
		return err
	}
	return Update(context.Background(), UpdateConfig{
		Auth:       auth,
		BaseURL:    c.OrchestrateBaseURL,
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
	return PsiphonQuery(ctx, PsiphonConfig{
		Auth:       auth,
		BaseURL:    c.OrchestrateBaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	})
}

// FetchTorTargets returns the targets for the tor experiment.
func (c *Client) FetchTorTargets(ctx context.Context) (map[string]model.TorTarget, error) {
	_, auth, err := c.getCredsAndAuth()
	if err != nil {
		return nil, err
	}
	return TorQuery(ctx, TorConfig{
		Auth:       auth,
		BaseURL:    c.OrchestrateBaseURL,
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
