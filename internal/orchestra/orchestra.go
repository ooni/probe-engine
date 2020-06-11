// Package orchestra contains common orchestra code. Orchestra is a set of
// OONI APIs for probe orchestration. You can find code implementing each
// specific API into a subpackage of this package. This package contains the
// toplevel orchestra client that the session should use.
package orchestra

import (
	"errors"
	"math/rand"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
)

// Client is a client for OONI orchestra
type Client struct {
	httpx.Client
	LoginCalls    *atomicx.Int64
	RegisterCalls *atomicx.Int64
	StateFile     *StateFile
}

// NewClient creates a new client.
func NewClient(sess model.ExperimentSession, endpoint model.Service) (*Client, error) {
	client := &Client{
		Client: httpx.Client{
			BaseURL:    endpoint.Address,
			HTTPClient: sess.DefaultHTTPClient(),
			Logger:     sess.Logger(),
			UserAgent:  sess.UserAgent(),
		},
		LoginCalls:    atomicx.NewInt64(),
		RegisterCalls: atomicx.NewInt64(),
		StateFile:     NewStateFile(sess.KeyValueStore()),
	}
	switch endpoint.Type {
	case "https":
		return client, nil
	default:
		return nil, errors.New("unsupported endpoint type")
	}
}

var (
	errInvalidMetadata = errors.New("invalid metadata")
	errNotLoggedIn     = errors.New("not logged in")
	errNotRegistered   = errors.New("not registered")
)

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
