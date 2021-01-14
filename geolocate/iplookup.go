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

var (
	// ErrAllIPLookuppersFailed indicates that we failed with looking
	// up the probe IP for with all the lookuppers that we tried.
	ErrAllIPLookuppersFailed = errors.New("all IP lookuppers failed")

	// ErrInvalidIPAddress indicates that the code returned to us a
	// string that actually isn't a valid IP address.
	ErrInvalidIPAddress = errors.New("lookupper did not return a valid IP")
)

type lookupFunc func(
	ctx context.Context, client *http.Client,
	logger model.Logger, userAgent string,
) (string, error)

type method struct {
	name string
	fn   lookupFunc
}

var (
	methods = []method{
		{
			name: "avast",
			fn:   avastIPLookup,
		},
		{
			name: "ipconfig",
			fn:   ipConfigIPLookup,
		},
		{
			name: "ipinfo",
			fn:   ipInfoIPLookup,
		},
		{
			name: "stun_ekiga",
			fn:   stunEkigaIPLookup,
		},
		{
			name: "stun_google",
			fn:   stunGoogleIPLookup,
		},
		{
			name: "ubuntu",
			fn:   ubuntuIPLookup,
		},
	}

	once sync.Once
)

type ipLookupClient struct {
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

func (c ipLookupClient) doWithCustomFunc(
	ctx context.Context, fn lookupFunc,
) (string, error) {
	ip, err := fn(ctx, c.HTTPClient, c.Logger, c.UserAgent)
	if err != nil {
		return model.DefaultProbeIP, err
	}
	if net.ParseIP(ip) == nil {
		return model.DefaultProbeIP, fmt.Errorf("%w: %s", ErrInvalidIPAddress, ip)
	}
	c.Logger.Debugf("iplookup: IP: %s", ip)
	return ip, nil
}

func (c ipLookupClient) LookupProbeIP(ctx context.Context) (string, error) {
	union := multierror.New(ErrAllIPLookuppersFailed)
	for _, method := range makeSlice() {
		c.Logger.Debugf("iplookup: using %s", method.name)
		ip, err := c.doWithCustomFunc(ctx, method.fn)
		if err == nil {
			return ip, nil
		}
		union.Add(err)
	}
	return model.DefaultProbeIP, union
}
