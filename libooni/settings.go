package main

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

// Options contains the settings options
type Options struct {
	// AssetsDir is the directory where to store assets
	AssetsDir string `json:"assets_dir"`

	// Backend is a test helper for a nettest
	Backend string `json:"backend,omitempty"`

	// BouncerBaseURL contains the bouncer base URL
	BouncerBaseURL string `json:"bouncer_base_url,omitempty"`

	// CaBundlePath contains the CA bundle path
	CaBundlePath string `json:"net/ca_bundle_path,omitempty"`

	// CollectorBaseURL contains the collector base URL
	CollectorBaseURL string `json:"collector_base_url,omitempty"`

	// DataDir is the directory where to store persitent data
	DataDir string `json:"data_dir"`

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

	// NoGeoIP indicates whether to perform a GeoIP lookup
	NoGeoIP bool `json:"no_geoip,omitempty"`

	// NoResolverLookup indicates whether to perform a resolver lookup
	NoResolverLookup bool `json:"no_resolver_lookup"`

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

	// TempDir is the temporary directory
	TempDir string `json:"temp_dir"`
}
