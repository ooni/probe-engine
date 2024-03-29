// Package dnsping is the experimental dnsping experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-035-dnsping.md.
package dnsping

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ooni/probe-engine/pkg/logx"
	"github.com/ooni/probe-engine/pkg/measurexlite"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

const (
	testName    = "dnsping"
	testVersion = "0.4.0"
)

// Config contains the experiment configuration.
type Config struct {
	// Delay is the delay between each repetition (in milliseconds).
	Delay int64 `ooni:"number of milliseconds to wait before sending each ping"`

	// Domains is the space-separated list of domains to measure.
	Domains string `ooni:"space-separated list of domains to measure"`

	// Repetitions is the number of repetitions for each ping.
	Repetitions int64 `ooni:"number of times to repeat the measurement"`
}

func (c *Config) delay() time.Duration {
	if c.Delay > 0 {
		return time.Duration(c.Delay) * time.Millisecond
	}
	return time.Second
}

func (c Config) repetitions() int64 {
	if c.Repetitions > 0 {
		return c.Repetitions
	}
	return 10
}

func (c Config) domains() string {
	if c.Domains != "" {
		return c.Domains
	}
	return "edge-chat.instagram.com example.com"
}

// Measurer performs the measurement.
type Measurer struct {
	config Config
}

// ExperimentName implements ExperimentMeasurer.ExperiExperimentName.
func (m *Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion implements ExperimentMeasurer.ExperimentVersion.
func (m *Measurer) ExperimentVersion() string {
	return testVersion
}

var (
	// errNoInputProvided indicates you didn't provide any input
	errNoInputProvided = errors.New("not input provided")

	// errInputIsNotAnURL indicates that input is not an URL
	errInputIsNotAnURL = errors.New("input is not an URL")

	// errInvalidScheme indicates that the scheme is invalid
	errInvalidScheme = errors.New("scheme must be udp")

	// errMissingPort indicates that there is no port.
	errMissingPort = errors.New("the URL must include a port")
)

// Run implements ExperimentMeasurer.Run.
func (m *Measurer) Run(ctx context.Context, args *model.ExperimentArgs) error {
	// unpack experiment args
	_ = args.Callbacks
	measurement := args.Measurement
	sess := args.Session
	if measurement.Input == "" {
		return errNoInputProvided
	}

	// parse experiment input
	parsed, err := url.Parse(string(measurement.Input))
	if err != nil {
		return fmt.Errorf("%w: %s", errInputIsNotAnURL, err.Error())
	}
	if parsed.Scheme != "udp" {
		return errInvalidScheme
	}
	if parsed.Port() == "" {
		return errMissingPort
	}

	// create the empty measurement test keys
	tk := NewTestKeys()
	measurement.TestKeys = tk

	// parse the domains to measure
	domains := strings.Split(m.config.domains(), " ")

	// spawn a pinger for each domain to measure
	wg := new(sync.WaitGroup)
	wg.Add(len(domains))
	for _, domain := range domains {
		go m.dnsPingLoop(ctx, measurement.MeasurementStartTimeSaved, sess.Logger(), parsed.Host, domain, wg, tk)
	}

	// block until all pingers are done
	wg.Wait()

	// generate textual summary
	summarize(tk)

	return nil // return nil so we always submit the measurement
}

// dnsPingLoop sends all the ping requests and emits the results onto the out channel.
func (m *Measurer) dnsPingLoop(ctx context.Context, zeroTime time.Time, logger model.Logger,
	address string, domain string, wg *sync.WaitGroup, tk *TestKeys) {
	// make sure the parent knows when we're done
	defer wg.Done()

	// create ticker so we know when to send the next DNS ping
	ticker := time.NewTicker(m.config.delay())
	defer ticker.Stop()

	// start a goroutine for each ping repetition
	for i := int64(0); i < m.config.repetitions(); i++ {
		wg.Add(1)
		go m.dnsRoundTrip(ctx, i, zeroTime, logger, address, domain, wg, tk)

		// make sure we wait until it's time to send the next ping
		<-ticker.C
	}
}

// dnsRoundTrip performs a round trip and returns the results to the caller.
func (m *Measurer) dnsRoundTrip(ctx context.Context, index int64, zeroTime time.Time,
	logger model.Logger, address string, domain string, wg *sync.WaitGroup, tk *TestKeys) {
	// create context bound to timeout
	// TODO(bassosimone): make the timeout user-configurable
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// make sure we inform the parent when we're done
	defer wg.Done()

	// create trace for collecting information
	trace := measurexlite.NewTrace(index, zeroTime)

	// create dialer and resolver
	//
	// TODO(bassosimone, DecFox): what should we do if the user passes us a resolver with a
	// domain name in terms of saving its results? Shall we save also the system resolver's lookups?
	// Shall we, otherwise, pre-resolve the domain name to IP addresses once and for all? In such
	// a case, shall we use all the available IP addresses or just some of them?
	dialer := netxlite.NewDialerWithStdlibResolver(logger)
	resolver := trace.NewParallelUDPResolver(logger, dialer, address)

	// perform the lookup proper
	ol := logx.NewOperationLogger(logger, "DNSPing #%d %s %s", index, address, domain)
	addrs, err := resolver.LookupHost(ctx, domain)
	stopOperationLogger(ol, addrs, err)

	// wait a bit for delayed responses
	delayedResps := trace.DelayedDNSResponseWithTimeout(ctx, 250*time.Millisecond)

	// assemble the results by inspecting ordinary and late responses
	pings := []*SinglePing{}
	for _, lookup := range trace.DNSLookupsFromRoundTrip() {
		// make sure we only include the query types we care about (in principle, there
		// should be no other query, so we're doing this just for robustness).
		if lookup.QueryType == "A" || lookup.QueryType == "AAAA" {
			sp := &SinglePing{
				Query:            lookup,
				DelayedResponses: []*model.ArchivalDNSLookupResult{},
			}

			// now take care of delayed responses
			if len(delayedResps) > 0 {
				logger.Warnf("DNSPing #%d... received %d delayed responses", index, len(delayedResps))
				// record the delayed responses of the corresponding query
				for _, resp := range delayedResps {
					if resp.QueryType == lookup.QueryType {
						sp.DelayedResponses = append(sp.DelayedResponses, resp)
					}
				}
			}

			pings = append(pings, sp)
		}
	}

	tk.addPings(pings)
}

type stoppableOperationLogger interface {
	Stop(value any)
}

func stopOperationLogger(ol stoppableOperationLogger, addrs []string, err error) {
	if err == nil {
		ol.Stop(strings.Join(addrs, " "))
	} else {
		ol.Stop(err)
	}
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &Measurer{config: config}
}
