// Package urlgetter implements a nettest that fetches a URL. This is not
// an official OONI nettest, but rather is a probe-engine specific internal
// experimental nettest that can be useful to do research.
package urlgetter

import (
	"context"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
)

const (
	testName    = "urlgetter"
	testVersion = "0.0.3"
)

// Config contains the experiment's configuration.
type Config struct {
	DNSCache          string `ooni:"Add 'DOMAIN IP...' to cache"`
	HTTPHost          string `ooni:"Force using specific HTTP Host header"`
	NoFollowRedirects bool   `ooni:"Disable following redirects"`
	NoTLSVerify       bool   `ooni:"Disable TLS verification"`
	RejectDNSBogons   bool   `ooni:"Fail DNS lookup if response contains bogons"`
	ResolverURL       string `ooni:"URL describing the resolver to use"`
	TLSServerName     string `ooni:"Force TLS to using a specific SNI in Client Hello"`
	TLSVersion        string `ooni:"Force specific TLS version (e.g. 'TLSv1.3')"`
	Tunnel            string `ooni:"Run experiment over a tunnel, e.g. psiphon"`
}

// TestKeys contains the experiment's result.
type TestKeys struct {
	Agent         string                     `json:"agent"`
	BootstrapTime float64                    `json:"bootstrap_time,omitempty"`
	DNSCache      []string                   `json:"dns_cache,omitempty"`
	Failure       *string                    `json:"failure"`
	NetworkEvents []archival.NetworkEvent    `json:"network_events"`
	Queries       []archival.DNSQueryEntry   `json:"queries"`
	Requests      []archival.RequestEntry    `json:"requests"`
	TCPConnect    []archival.TCPConnectEntry `json:"tcp_connect"`
	SOCKSProxy    string                     `json:"socksproxy,omitempty"`
	TLSHandshakes []archival.TLSHandshake    `json:"tls_handshakes"`
	Tunnel        string                     `json:"tunnel,omitempty"`
}

// RegisterExtensions registers the extensions used by the urlgetter
// experiment into the provided measurement.
func RegisterExtensions(m *model.Measurement) {
	archival.ExtHTTP.AddTo(m)
	archival.ExtDNS.AddTo(m)
	archival.ExtNetevents.AddTo(m)
	archival.ExtTLSHandshake.AddTo(m)
	archival.ExtTunnel.AddTo(m)
}

type measurer struct {
	Config
}

func (m measurer) ExperimentName() string {
	return testName
}

func (m measurer) ExperimentVersion() string {
	return testVersion
}

func (m measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	RegisterExtensions(measurement)
	g := Getter{
		Config:  m.Config,
		Session: sess,
		Target:  string(measurement.Input),
	}
	tk, err := g.Get(ctx)
	measurement.TestKeys = tk
	return err
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return measurer{Config: config}
}
