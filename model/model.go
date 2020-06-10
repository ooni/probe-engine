// Package model defines shared data structures.
package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// MeasurementTarget is the target of a OONI measurement.
type MeasurementTarget string

// MarshalJSON serializes the MeasurementTarget.
func (t MeasurementTarget) MarshalJSON() ([]byte, error) {
	if t == "" {
		return json.Marshal(nil)
	}
	return json.Marshal(string(t))
}

// Measurement is a OONI measurement.
//
// This structure is compatible with the definition of the base data format in
// https://github.com/ooni/spec/blob/master/data-formats/df-000-base.md.
type Measurement struct {
	// Annotations contains results annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// DataFormatVersion is the version of the data format
	DataFormatVersion string `json:"data_format_version"`

	// Extensions contains information about the extensions included
	// into the test_keys of this measurement.
	Extensions map[string]int64 `json:"extensions,omitempty"`

	// ID is the locally generated measurement ID
	ID string `json:"id,omitempty"`

	// Input is the measurement input
	Input MeasurementTarget `json:"input"`

	// InputHashes contains input hashes
	InputHashes []string `json:"input_hashes,omitempty"`

	// MeasurementStartTime is the time when the measurement started
	MeasurementStartTime string `json:"measurement_start_time"`

	// MeasurementStartTimeSaved is the moment in time when we
	// started the measurement. This is not included into the JSON
	// and is only used within probe-engine as a "zero" time.
	MeasurementStartTimeSaved time.Time `json:"-"`

	// OOID is the measurement ID stamped by the OONI collector.
	OOID string `json:"ooid,omitempty"`

	// Options contains command line options
	Options []string `json:"options,omitempty"`

	// ProbeASN contains the probe autonomous system number
	ProbeASN string `json:"probe_asn"`

	// ProbeCC contains the probe country code
	ProbeCC string `json:"probe_cc"`

	// ProbeCity contains the probe city
	ProbeCity string `json:"probe_city,omitempty"`

	// ProbeIP contains the probe IP
	ProbeIP string `json:"probe_ip,omitempty"`

	// ProbeNetworkName contains the probe network name
	ProbeNetworkName string `json:"probe_network_name,omitempty"`

	// ReportID contains the report ID
	ReportID string `json:"report_id"`

	// ResolverASN is the ASN of the resolver
	ResolverASN string `json:"resolver_asn"`

	// ResolverIP is the resolver IP
	ResolverIP string `json:"resolver_ip"`

	// ResolverNetworkName is the network name of the resolver.
	ResolverNetworkName string `json:"resolver_network_name"`

	// SoftwareName contains the software name
	SoftwareName string `json:"software_name"`

	// SoftwareVersion contains the software version
	SoftwareVersion string `json:"software_version"`

	// TestHelpers contains the test helpers. It seems this structure is more
	// complex than we would like. In particular, using a map from string to
	// string does not fit into the web_connectivity use case. Hence, for now
	// we're going to represent this using interface{}. In going forward we
	// may probably want to have more uniform test helpers.
	TestHelpers map[string]interface{} `json:"test_helpers,omitempty"`

	// TestKeys contains the real test result. This field is opaque because
	// each experiment will insert here a different structure.
	TestKeys interface{} `json:"test_keys"`

	// TestName contains the test name
	TestName string `json:"test_name"`

	// MeasurementRuntime contains the measurement runtime. The JSON name
	// is test_runtime because this is the name expected by the OONI backend
	// even though that name is clearly a misleading one.
	MeasurementRuntime float64 `json:"test_runtime"`

	// TestStartTime contains the test start time
	TestStartTime string `json:"test_start_time"`

	// TestVersion contains the test version
	TestVersion string `json:"test_version"`
}

// AddAnnotations adds the annotations from input to m.Annotations.
func (m *Measurement) AddAnnotations(input map[string]string) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	for key, value := range input {
		m.Annotations[key] = value
	}
}

// AddAnnotation adds a single annotations to m.Annotations.
func (m *Measurement) AddAnnotation(key, value string) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	m.Annotations[key] = value
}

// MakeGenericTestKeys casts the m.TestKeys to a map[string]interface{}.
//
// Ideally, all tests should have a clear Go structure, well defined, that
// will be stored in m.TestKeys as an interface. This is not already the
// case and it's just valid for tests written in Go. Until all tests will
// be written in Go, we'll keep this glue here to make sure we convert from
// the engine format to the cli format.
//
// This function will first attempt to cast directly to map[string]interface{},
// which is possible for MK tests, and then use JSON serialization and
// de-serialization only if that's required.
func (m *Measurement) MakeGenericTestKeys() (map[string]interface{}, error) {
	return m.makeGenericTestKeys(json.Marshal)
}

func (m *Measurement) makeGenericTestKeys(
	marshal func(v interface{}) ([]byte, error),
) (map[string]interface{}, error) {
	if result, ok := m.TestKeys.(map[string]interface{}); ok {
		return result, nil
	}
	data, err := marshal(m.TestKeys)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

// Service describes a backend service.
//
// The fields of this struct have the meaning described in v2.0.0 of the OONI
// bouncer specification defined by
// https://github.com/ooni/spec/blob/master/backends/bk-004-bouncer.md.
type Service struct {
	// Address is the address of the server.
	Address string `json:"address"`

	// Type is the type of the service.
	Type string `json:"type"`

	// Front is the front to use with "cloudfront" type entries.
	Front string `json:"front,omitempty"`
}

// LocationInfo contains location information
type LocationInfo struct {
	// ASN is the autonomous system number
	ASN uint

	// CountryCode is the country code
	CountryCode string

	// NetworkName is the network name
	NetworkName string

	// IP is the probe IP
	ProbeIP string

	// ResolverASN is the resolver ASN
	ResolverASN uint

	// ResolverIP is the resolver IP
	ResolverIP string

	// ResolverNetworkName is the resolver network name
	ResolverNetworkName string
}

// PrivacySettings contains privacy settings for submitting measurements.
type PrivacySettings struct {
	// IncludeASN indicates whether to include the ASN
	IncludeASN bool

	// IncludeCountry indicates whether to include the country
	IncludeCountry bool

	// IncludeIP indicates whether to include the IP
	IncludeIP bool
}

// Apply applies the privacy settings to the measurement, possibly
// scrubbing the probeIP out of it.
func (ps PrivacySettings) Apply(m *Measurement, probeIP string) (err error) {
	if ps.IncludeASN == false {
		m.ProbeASN = DefaultProbeASNString
	}
	if ps.IncludeCountry == false {
		m.ProbeCC = DefaultProbeCC
	}
	if ps.IncludeIP == false {
		m.ProbeIP = DefaultProbeIP
		err = ps.MaybeRewriteTestKeys(m, probeIP, json.Marshal)
	}
	return
}

// MaybeRewriteTestKeys is the function called by Apply that
// ensures that m's serialization doesn't include the IP
func (ps PrivacySettings) MaybeRewriteTestKeys(
	m *Measurement, currentIP string,
	marshal func(interface{}) ([]byte, error),
) error {
	if net.ParseIP(currentIP) == nil {
		return errors.New("Invalid probe IP string")
	}
	data, err := marshal(m.TestKeys)
	if err != nil {
		return err
	}
	// The check using Count is to save an unnecessary copy performed by
	// ReplaceAll when there are no matches into the body. This is what
	// we would like the common case to be, meaning that the code has done
	// its job correctly and has not leaked the IP.
	bpip := []byte(currentIP)
	if bytes.Count(data, bpip) <= 0 {
		return nil
	}
	data = bytes.ReplaceAll(data, bpip, []byte(`[REDACTED]`))
	// We add an annotation such that hopefully later we can measure the
	// number of cases where we failed to sanitize properly.
	m.AddAnnotation("_probe_engine_sanitize_test_keys", "true")
	return json.Unmarshal(data, &m.TestKeys)
}

const (
	// DefaultProbeASN is the default probe ASN as number.
	DefaultProbeASN uint = 0

	// DefaultProbeCC is the default probe CC.
	DefaultProbeCC = "ZZ"

	// DefaultProbeIP is the default probe IP.
	DefaultProbeIP = "127.0.0.1"

	// DefaultProbeNetworkName is the default probe network name.
	DefaultProbeNetworkName = ""

	// DefaultResolverASN is the default resolver ASN.
	DefaultResolverASN uint = 0

	// DefaultResolverIP is the default resolver IP.
	DefaultResolverIP = "127.0.0.2"

	// DefaultResolverNetworkName is the default resolver network name.
	DefaultResolverNetworkName = ""
)

var (
	// DefaultProbeASNString is the default probe ASN as a string.
	DefaultProbeASNString = fmt.Sprintf("AS%d", DefaultProbeASN)

	// DefaultResolverASNString is the default resolver ASN as a string.
	DefaultResolverASNString = fmt.Sprintf("AS%d", DefaultResolverASN)
)

// URLInfo contains info on a test lists URL
type URLInfo struct {
	CategoryCode string `json:"category_code"`
	CountryCode  string `json:"country_code"`
	URL          string `json:"url"`
}

// KeyValueStore is a key-value store used by the session.
type KeyValueStore interface {
	Get(key string) (value []byte, err error)
	Set(key string, value []byte) (err error)
}

// TorTarget is a target for the tor experiment.
type TorTarget struct {
	// Address is the address of the target.
	Address string `json:"address"`

	// Name is the name of the target.
	Name string `json:"name"`

	// Params contains optional params for, e.g., pluggable transports.
	Params map[string][]string `json:"params"`

	// Protocol is the protocol to use with the target.
	Protocol string `json:"protocol"`
}

// Logger defines the common interface that a logger should have. It is
// out of the box compatible with `log.Log` in `apex/log`.
type Logger interface {
	// Debug emits a debug message.
	Debug(msg string)

	// Debugf formats and emits a debug message.
	Debugf(format string, v ...interface{})

	// Info emits an informational message.
	Info(msg string)

	// Infof format and emits an informational message.
	Infof(format string, v ...interface{})

	// Warn emits a warning message.
	Warn(msg string)

	// Warnf formats and emits a warning message.
	Warnf(format string, v ...interface{})
}

// ExperimentOrchestraClient is the experiment's view of
// a client for querying the OONI orchestra.
type ExperimentOrchestraClient interface {
	FetchPsiphonConfig(ctx context.Context) ([]byte, error)
	FetchTorTargets(ctx context.Context) (map[string]TorTarget, error)
}

// ExperimentSession is the experiment's view of a session.
type ExperimentSession interface {
	ASNDatabasePath() string
	CABundlePath() string
	GetTestHelpersByName(name string) ([]Service, bool)
	DefaultHTTPClient() *http.Client
	Logger() Logger
	MaybeStartTunnel(ctx context.Context, name string) error
	NewOrchestraClient(ctx context.Context) (ExperimentOrchestraClient, error)
	KeyValueStore() KeyValueStore
	ProbeASNString() string
	ProbeCC() string
	ProbeIP() string
	ProbeNetworkName() string
	ProxyURL() *url.URL
	SoftwareName() string
	SoftwareVersion() string
	TempDir() string
	TorArgs() []string
	TorBinary() string
	TunnelBootstrapTime() time.Duration
	UserAgent() string
}

// ExperimentCallbacks contains experiment event-handling callbacks
type ExperimentCallbacks interface {
	// OnDataUsage provides information about data usage.
	//
	// This callback is deprecated and will be removed once we have
	// removed the dependency on Measurement Kit.
	OnDataUsage(dloadKiB, uploadKiB float64)

	// OnProgress provides information about an experiment progress.
	OnProgress(percentage float64, message string)
}

// ExperimentMeasurer is the interface that allows to run a
// measurement for a specific experiment.
type ExperimentMeasurer interface {
	// ExperimentName returns the experiment name.
	ExperimentName() string

	// ExperimentVersion returns the experiment version.
	ExperimentVersion() string

	// Run runs the experiment with the specified context, session,
	// measurement, and experiment calbacks.
	Run(
		ctx context.Context, sess ExperimentSession,
		measurement *Measurement, callbacks ExperimentCallbacks,
	) error
}
