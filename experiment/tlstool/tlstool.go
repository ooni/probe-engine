// Package tlstool contains a TLS tool that we are currently using
// for running quick and dirty experiments. This tool will change
// without notice and may be removed without notice.
package tlstool

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/ooni/probe-engine/experiment/tlstool/internal"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx"
	"github.com/ooni/probe-engine/netx/archival"
)

const (
	testName    = "tlstool"
	testVersion = "0.0.1"
)

// Config contains the experiment configuration.
type Config struct {
	Delay int64  `ooni:"Milliseconds to wait between writes"`
	SNI   string `ooni:"Set the specified SNI"`
}

// TestKeys contains the experiment results.
type TestKeys struct {
	VanillaFailure *string `json:"vanilla_failure"`
	SplitFailure   *string `json:"split_failure"`
}

// Measurer performs the measurement.
type Measurer struct {
	config Config
}

// ExperimentName implements ExperimentMeasurer.ExperiExperimentName.
func (m Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion implements ExperimentMeasurer.ExperimentVersion.
func (m Measurer) ExperimentVersion() string {
	return testVersion
}

// Run implements ExperimentMeasurer.Run.
func (m Measurer) Run(
	ctx context.Context,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
) error {
	tk := new(TestKeys)
	measurement.TestKeys = tk
	address := string(measurement.Input)
	err := m.vanillaRun(ctx, sess.Logger(), address)
	callbacks.OnProgress(0.5, fmt.Sprintf("vanilla: %+v", err))
	tk.VanillaFailure = archival.NewFailure(err)
	err = m.splitRun(ctx, sess.Logger(), address)
	callbacks.OnProgress(1.0, fmt.Sprintf("split: %+v", err))
	tk.SplitFailure = archival.NewFailure(err)
	return nil
}

func (m Measurer) newDialer(logger model.Logger) netx.Dialer {
	resolver, err := netx.NewDNSClientWithOverrides(netx.Config{Logger: logger},
		"https://cloudflare.com/dns-query", "dns.cloudflare.com", "")
	runtimex.PanicOnError(err, "cannot initialize resolver")
	return netx.NewDialer(netx.Config{FullResolver: resolver, Logger: logger})
}

func (m Measurer) vanillaRun(ctx context.Context, logger model.Logger, address string) error {
	dialer := m.newDialer(logger)
	tdialer := netx.NewTLSDialer(netx.Config{
		Dialer:    dialer,
		Logger:    logger,
		TLSConfig: m.tlsConfig(),
	})
	conn, err := tdialer.DialTLSContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (m Measurer) splitRun(ctx context.Context, logger model.Logger, address string) error {
	dialer := &internal.SplitDialer{
		Dialer:  m.newDialer(logger),
		Delay:   m.config.Delay,
		Pattern: m.pattern(address),
	}
	tdialer := netx.NewTLSDialer(netx.Config{
		Dialer:    dialer,
		Logger:    logger,
		TLSConfig: m.tlsConfig(),
	})
	conn, err := tdialer.DialTLSContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (m Measurer) tlsConfig() *tls.Config {
	if m.config.SNI != "" {
		return &tls.Config{ServerName: m.config.SNI}
	}
	return nil
}

func (m Measurer) pattern(address string) string {
	if m.config.SNI != "" {
		return m.config.SNI
	}
	addr, _, err := net.SplitHostPort(address)
	runtimex.PanicOnError(err, "cannot split address")
	return addr
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return Measurer{config: config}
}
