// Package doh implements OONI's DoH experiment.
//
// Spec: https://github.com/ooni/spec/blob/master/nettests/ts-023-doh.md
package doh

import (
	"context"
	"sync"
	"time"

	"github.com/ooni/netx"
	"github.com/ooni/netx/dnsx"
	"github.com/ooni/netx/handlers"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "doh"
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct {
	// URL is the DoH URL to use. If not set, we will use the
	// default URL indicated in the spec.
	URL string
}

// TestKeys contains the experiment test keys
type TestKeys struct {
	Failure    string   `json:"failure"`
	URL        string   `json:"url"`
	XAddresses []string `json:"x-addresses"`
}

type measurer struct {
	mutex    sync.Mutex
	resolver dnsx.Client
}

func (m *measurer) initIdempotent(config *Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.resolver != nil {
		return nil
	}
	if config.URL == "" {
		config.URL = "https://mozilla.cloudflare-dns.com/dns-query"
	}
	dialer := netx.NewDialer(handlers.StdoutHandler)
	resolver, err := dialer.NewResolver("doh", config.URL)
	if err != nil {
		return err
	}
	m.resolver = resolver
	return nil
}

func (m *measurer) do(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tk := &TestKeys{}
	measurement.TestKeys = tk
	err := m.initIdempotent(&config)
	if err != nil {
		tk.Failure = err.Error()
		return err
	}
	tk.URL = config.URL // must be after initIdempotent
	addrs, err := m.resolver.LookupHost(ctx, measurement.Input)
	if err != nil {
		tk.Failure = err.Error()
		return err
	}
	tk.XAddresses = addrs
	return nil
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	m := &measurer{}
	return experiment.New(
		sess, testName, testVersion,
		func(
			ctx context.Context,
			sess *session.Session,
			measurement *model.Measurement,
			callbacks handler.Callbacks,
		) error {
			return m.do(ctx, sess, measurement, callbacks, config)
		})
}
