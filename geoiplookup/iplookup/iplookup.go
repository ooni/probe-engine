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

	"github.com/ooni/probe-engine/geoiplookup/constants"
	"github.com/ooni/probe-engine/geoiplookup/iplookup/akamai"
	"github.com/ooni/probe-engine/geoiplookup/iplookup/avast"
	"github.com/ooni/probe-engine/geoiplookup/iplookup/ubuntu"
	"github.com/ooni/probe-engine/log"
)

type lookupFunc func(
	ctx context.Context, client *http.Client,
	logger log.Logger, userAgent string,
) (string, error)

type method struct {
	name string
	fn   lookupFunc
}

var (
	methods = []method{
		method{
			name: "akamai",
			fn:   akamai.Do,
		},
		method{
			name: "avast",
			fn:   avast.Do,
		},
		method{
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
	Logger log.Logger

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

func (c *Client) do(ctx context.Context, fn lookupFunc) (string, error) {
	ip, err := fn(ctx, c.HTTPClient, c.Logger, c.UserAgent)
	if err != nil {
		return constants.DefaultProbeIP, err
	}
	if net.ParseIP(ip) == nil {
		return constants.DefaultProbeIP, fmt.Errorf("Invalid IP address: %s", ip)
	}
	c.Logger.Debugf("iplookup: IP: %s", ip)
	return ip, nil
}

// Do performs the IP lookup.
func (c *Client) Do(ctx context.Context) (ip string, err error) {
	for _, method := range makeSlice() {
		c.Logger.Debugf("iplookup: using %s", method.name)
		ip, err = c.do(ctx, method.fn)
		if err == nil {
			return
		}
	}
	return constants.DefaultProbeIP, errors.New("All IP lookuppers failed")
}
