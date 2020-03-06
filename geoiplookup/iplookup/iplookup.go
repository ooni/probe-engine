// Package iplookup implements probe IP lookup.
package iplookup

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ooni/probe-engine/geoiplookup/iplookup/avast"
	"github.com/ooni/probe-engine/geoiplookup/iplookup/ubuntu"
	"github.com/ooni/probe-engine/model"
)

// LookupFunc is a function for performing the IP lookup.
type LookupFunc func(
	ctx context.Context, client *http.Client,
	logger model.Logger, userAgent string,
) (string, error)

type method struct {
	name string
	fn   LookupFunc
}

var (
	methods = []method{
		{
			name: "avast",
			fn:   avast.Do,
		},
		{
			name: "ubuntu",
			fn:   ubuntu.Do,
		},
	}

	once sync.Once
)

// Client is an iplookup client
type Client struct {
	// HTTPClient is the HTTP client to use
	HTTPClient *http.Client

	// Logger is the logger to use
	Logger model.Logger

	// UserAgent is the user agent to use
	UserAgent string
}

func makeSlice() []method {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := make([]method, len(methods))
	perm := r.Perm(len(methods))
	for idx, randIdx := range perm {
		ret[idx] = methods[randIdx]
	}
	return ret
}

// DoWithCustomFunc performs the IP lookup with a custom function.
func (c *Client) DoWithCustomFunc(
	ctx context.Context, fn LookupFunc,
) (string, error) {
	ip, err := fn(ctx, c.HTTPClient, c.Logger, c.UserAgent)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	if net.ParseIP(ip) == nil {
		return model.DefaultProbeIP, fmt.Errorf("Invalid IP address: %s", ip)
	}
	c.Logger.Debugf("iplookup: IP: %s", ip)
	return ip, nil
}

// Do performs the IP lookup.
func (c *Client) Do(ctx context.Context) (ip string, err error) {
	for _, method := range makeSlice() {
		c.Logger.Debugf("iplookup: using %s", method.name)
		ip, err = c.DoWithCustomFunc(ctx, method.fn)
		if err == nil {
			return
		}
	}
	return model.DefaultProbeIP, errors.New("All IP lookuppers failed")
}
