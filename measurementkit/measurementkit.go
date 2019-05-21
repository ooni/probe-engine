// Package measurementkit allows to use Measurement Kit.
package measurementkit

import (
	"encoding/json"

	"github.com/ooni/probe-engine/log"
)

// Settings contains settings
type Settings struct {
	// Annotations contains the annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// DisabledEvents contains disabled events
	DisabledEvents []string `json:"disabled_events,omitempty"`

	// Inputs contains the inputs
	Inputs []string `json:"inputs,omitempty"`

	// InputFilepaths contains the input file paths
	InputFilepaths []string `json:"input_filepaths,omitempty"`

	// LogLevel contains the logs level
	LogLevel string `json:"log_level,omitempty"`

	// Name contains the task name
	Name string `json:"name"`

	// Options contains the options
	Options Options `json:"options"`

	// OutputFilepath contains the output filepath
	OutputFilePath string `json:"output_filepath,omitempty"`
}

// Options contains the options
type Options struct {
	// Backend is a test helper for a nettest
	Backend string `json:"backend,omitempty"`

	// BouncerBaseURL contains the bouncer base URL
	BouncerBaseURL string `json:"bouncer_base_url,omitempty"`

	// CaBundlePath contains the CA bundle path
	CaBundlePath string `json:"net/ca_bundle_path,omitempty"`

	// CollectorBaseURL contains the collector base URL
	CollectorBaseURL string `json:"collector_base_url,omitempty"`

	// GeoIPCountryPath is the country database path
	GeoIPCountryPath string `json:"geoip_country_path,omitempty"`

	// GeoIPASNPath is the ASN database path
	GeoIPASNPath string `json:"geoip_asn_path,omitempty"`

	// MaxRuntime is the maximum runtime
	MaxRuntime float32 `json:"max_runtime,omitempty"`

	// NoBouncer indicates whether to use a bouncer
	NoBouncer bool `json:"no_bouncer,omitempty"`

	// NoCollector indicates whether to use a collector
	NoCollector bool `json:"no_collector,omitempty"`

	// NoFileReport indicates whether to write a report file
	NoFileReport bool `json:"no_file_report,omitempty"`

	// ProbeASN is the AS number
	ProbeASN string `json:"probe_asn,omitempty"`

	// ProbeCC is the probe country code
	ProbeCC string `json:"probe_cc,omitempty"`

	// ProbeIP is the probe IP
	ProbeIP string `json:"probe_ip,omitempty"`

	// ProbeNetworkName is the probe network name
	ProbeNetworkName string `json:"probe_network_name,omitempty"`

	// RandomizeInput indicates whether to randomize inputs
	RandomizeInput bool `json:"randomize_input,omitempty"`

	// SaveRealProbeIP indicates whether to save the real probe IP
	SaveRealProbeIP bool `json:"save_real_probe_ip,omitempty"`

	// SaveRealProbeIP indicates whether to save the real probe ASN
	SaveRealProbeASN bool `json:"save_real_probe_asn,omitempty"`

	// SaveRealProbeCC indicates whether to save the real probe CC
	SaveRealProbeCC bool `json:"save_real_probe_cc,omitempty"`

	// SoftwareName is the software name
	SoftwareName string `json:"software_name,omitempty"`

	// SoftwareVersion is the software version
	SoftwareVersion string `json:"software_version,omitempty"`
}

// NewSettings creates new Settings
func NewSettings(
	taskName string,
	softwareName string,
	softwareVersion string,
	caBundlePath string,
	probeASN string,
	probeCC string,
	probeIP string,
	probeNetworkName string,
) Settings {
	return Settings{
		LogLevel: "INFO",
		Name:     taskName,
		Options: Options{
			CaBundlePath:     caBundlePath,
			NoBouncer:        true,
			NoCollector:      true,
			NoFileReport:     true,
			ProbeASN:         probeASN,
			ProbeCC:          probeCC,
			ProbeIP:          probeIP,
			ProbeNetworkName: probeNetworkName,
			SaveRealProbeIP:  false,
			SoftwareName:     softwareName,
			SoftwareVersion:  softwareVersion,
		},
	}
}

// EventValue are all the possible value keys
type EventValue struct {
	// DownloadedKB is the amount of downloaded KiBs
	DownloadedKB float64 `json:"downloaded_kb,omitempty"`

	// Failure is the failure that occurred
	Failure string `json:"failure,omitempty"`

	// Idx is the measurement index
	Idx int64 `json:"idx,omitempty"`

	// Input is the input to which this event is related
	Input string `json:"input,omitempty"`

	// JSONStr is a serialized measurement
	JSONStr string `json:"json_str,omitempty"`

	// LogLevel is the log level
	LogLevel string `json:"log_level,omitempty"`

	// Message is the log message
	Message string `json:"message,omitempty"`

	// Percentage is the task progress
	Percentage float64 `json:"percentage,omitempty"`

	// ProbeASN is the probe ASN
	ProbeASN string `json:"probe_asn,omitempty"`

	// ProbeCC is the probe CC
	ProbeCC string `json:"probe_cc,omitempty"`

	// ProbeIP is the probe IP
	ProbeIP string `json:"probe_ip,omitempty"`

	// ProbeNetworkName is the probe network name
	ProbeNetworkName string `json:"probe_network_name,omitempty"`

	// ReportID is the report ID
	ReportID string `json:"report_id,omitempty"`

	// UploadedKB is the amount of uploaded KiBs
	UploadedKB float64 `json:"uploaded_kb,omitempty"`
}

// Event is a Measurement Kit event
type Event struct {
	// Is the key for the event
	Key string `json:"key"`

	// Contains the value for the event
	Value EventValue `json:"value"`
}

func loopEx(in <-chan []byte, out chan<- Event, logger log.Logger) {
	defer close(out)
	for data := range in {
		// Uncomment the following line to debug
		//logger.Debugf("measurementkit: event: %s", string(data))
		var event Event
		err := json.Unmarshal(data, &event)
		if err != nil {
			logger.Debugf("measurementkit: JSON processing error: %s", err.Error())
			continue
		}
		out <- event
	}
}

// StartEx is a more advanced Start that takes input settings
// and that emits Event on the returned channel.
func StartEx(settings Settings, logger log.Logger) (<-chan Event, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	logger.Debugf("measurementkit: settings: %s", string(data))
	in, err := Start(data)
	if err != nil {
		return nil, err
	}
	out := make(chan Event)
	go loopEx(in, out, logger)
	return out, nil
}

// Start starts a Measurement Kit task with the provided settings and
// returns a channel where events are emitted or an error.
func Start(settings []byte) (<-chan []byte, error) {
	return start(settings)
}

// IsAvailable indicates whether Measurement Kit support is available.
func IsAvailable() bool {
	return isAvailable()
}
