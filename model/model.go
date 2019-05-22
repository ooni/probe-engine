// Package model defines shared data structures.
package model

// Measurement is a OONI measurement.
//
// This structure is compatible with the definition of the base data format in
// https://github.com/ooni/spec/blob/master/data-formats/df-000-base.md.
type Measurement struct {
	// Annotations contains results annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// DataFormatVersion is the version of the data format
	DataFormatVersion string `json:"data_format_version"`

	// ID is the locally generated measurement ID
	ID string `json:"id,omitempty"`

	// Input is the measurement input
	Input string `json:"input,omitempty"`

	// InputHashes contains input hashes
	InputHashes []string `json:"input_hashes,omitempty"`

	// MeasurementStartTime is the time when the measurement started
	MeasurementStartTime string `json:"measurement_start_time"`

	// MeasurementRuntime contains the measurement runtime. The JSON name
	// is test_runtime because this is the name expected by the OONI backend
	// even though that name is clearly a misleading one.
	MeasurementRuntime float64 `json:"test_runtime"`

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

	// ReportID contains the report ID
	ReportID string `json:"report_id"`

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

	// TestStartTime contains the test start time
	TestStartTime string `json:"test_start_time"`

	// TestVersion contains the test version
	TestVersion string `json:"test_version"`
}

// AddAnnotations adds the annotations from input to m.Annotations.
func (m Measurement) AddAnnotations(input map[string]string) {
	for key, value := range input {
		m.Annotations[key] = value
	}
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

	// ResolverIP is the resolver IP
	ResolverIP string
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

	// DefaultResolverIP is the default resolver IP.
	DefaultResolverIP = "127.0.0.1"
)
