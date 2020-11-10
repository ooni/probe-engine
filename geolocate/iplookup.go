package geolocate

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ooni/probe-engine/internal/multierror"
	"github.com/ooni/probe-engine/model"
)

// ErrAllIPLookuppersFailed indicates that we failed with looking
// up the probe IP for with all the lookuppers that we tried.
var ErrAllIPLookuppersFailed = errors.New("all IP lookuppers failed")

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
			fn:   AvastIPLookup,
		},
		{
			name: "ipconfig",
			fn:   IPConfigIPLookup,
		},
		{
			name: "ipinfo",
			fn:   IPInfoIPLookup,
		},
		{
			name: "stun_ekiga",
			fn:   STUNEkigaIPLookup,
		},
		{
			name: "stun_google",
			fn:   STUNGoogleIPLookup,
		},
		{
			name: "ubuntu",
			fn:   UbuntuIPLookup,
		},
	}

	once sync.Once
)

// IPLookupClient is an iplookup client
type IPLookupClient struct {
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
func (c IPLookupClient) DoWithCustomFunc(
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
func (c IPLookupClient) Do(ctx context.Context) (string, error) {
	union := multierror.New(ErrAllIPLookuppersFailed)
	for _, method := range makeSlice() {
		c.Logger.Debugf("iplookup: using %s", method.name)
		ip, err := c.DoWithCustomFunc(ctx, method.fn)
		if err == nil {
			return ip, nil
		}
		union.Add(err)
	}
	return model.DefaultProbeIP, union
}
