// Package orchestra contains common orchestra code. Orchestra is a set of
// OONI APIs for probe orchestration. You can find code implementing each
// specific API into a subpackage of this package. This package contains the
// toplevel orchestra client that the session should use.
package orchestra

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/model"
)

// Client is a client for OONI orchestra
type Client struct {
	BaseURL       string
	HTTPClient    *http.Client
	Logger        model.Logger
	LoginCalls    *atomicx.Int64
	RegisterCalls *atomicx.Int64
	StateFile     *StateFile
	UserAgent     string
}

// NewClient creates a new client.
func NewClient(
	httpClient *http.Client, logger model.Logger,
	userAgent string, stateFile *StateFile,
) *Client {
	return &Client{
		BaseURL:       "https://ps.ooni.io",
		HTTPClient:    httpClient,
		Logger:        logger,
		LoginCalls:    atomicx.NewInt64(),
		RegisterCalls: atomicx.NewInt64(),
		StateFile:     stateFile,
		UserAgent:     userAgent,
	}
}

var (
	errInvalidMetadata = errors.New("invalid metadata")
	errNotLoggedIn     = errors.New("not logged in")
	errNotRegistered   = errors.New("not registered")
)

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
