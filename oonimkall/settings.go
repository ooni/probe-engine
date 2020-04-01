package oonimkall

// settingsRecord contains settings for a task. This structure extends the one
// described by MK v0.10.9 FFI API (https://git.io/Jv4Rv).
type settingsRecord struct {
	// Annotations contains the annotations to be added
	// to every measurements performed by the task.
	Annotations map[string]string `json:"annotations,omitempty"`

	// AssetsDir is the directory where to store assets. This
	// field is an extension of MK's specification. If
	// this field is empty, the task won't start.
	AssetsDir string `json:"assets_dir"`

	// DisabledEvents contains disabled events. See
	// https://git.io/Jv4Rv for the events names.
	DisabledEvents []string `json:"disabled_events,omitempty"`

	// Inputs contains the inputs. The task will fail if it
	// requires input and you provide no input.
	Inputs []string `json:"inputs,omitempty"`

	// InputFilepaths contains the input file paths. This
	// setting is not implemented by this library. Attempting
	// to set it will cause a startup error.
	InputFilepaths []string `json:"input_filepaths,omitempty"`

	// LogLevel contains the logs level. See https://git.io/Jv4Rv
	// for the names of the available log levels.
	LogLevel string `json:"log_level,omitempty"`

	// Name contains the task name. By https://git.io/Jv4Rv the
	// names are in camel case, e.g. `Ndt`.
	Name string `json:"name"`

	// Options contains the task options.
	Options settingsOptions `json:"options"`

	// OutputFilepath contains the output filepath. This
	// setting is not implemented by this library. Attempting
	// to set it will cause a startup error unless the
	// Options.NoFileReport setting is true.
	OutputFilepath string `json:"output_filepath,omitempty"`

	// StateDir is the directory where to store persistent data. This
	// field is an extension of MK's specification. If
	// this field is empty, the task won't start.
	StateDir string `json:"state_dir"`

	// TempDir is the temporary directory. This
	// field is an extension of MK's specification. If
	// this field is empty, the task won't start.
	TempDir string `json:"temp_dir"`
}

// settingsOptions contains the settings options
type settingsOptions struct {
	// AllEndpoints is a WhatsApp specific option indicating that we
	// should test all endpoints rather than a random susbet. This
	// library does not support this setting and fails if you provide
	// it as input.
	AllEndpoints *bool `json:"all_endpoints,omitempty"`

	// Backend is a test helper for a nettest. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	Backend string `json:"backend,omitempty"`

	// BouncerBaseURL contains the bouncer base URL
	BouncerBaseURL string `json:"bouncer_base_url,omitempty"`

	// CABundlePath contains the CA bundle path. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup warning, and the
	// library will otherwise ignore this setting.
	CABundlePath string `json:"net/ca_bundle_path,omitempty"`

	// CollectorBaseURL contains the collector base URL
	CollectorBaseURL string `json:"collector_base_url,omitempty"`

	// ConstantBitrate was an option for the DASH experiment that
	// this library does not support. Setting it to any value will
	// cause the code to stop early with a startup failure.
	ConstantBitrate *bool `json:"constant_bitrate,omitempty"`

	// DNSNameserver is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	DNSNameserver *string `json:"dns_nameserver,omitempty"`

	// DNSEngine is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	DNSEngine *string `json:"dns_engine,omitempty"`

	// ExpectedBody is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	ExpectedBody *string `json:"expected_body,omitempty"`

	// GeoIPASNPath is the ASN database path. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup warning, and the
	// library will otherwise ignore this setting.
	GeoIPASNPath string `json:"geoip_asn_path,omitempty"`

	// GeoIPCountryPath is the country database path. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup warning, and the
	// library will otherwise ignore this setting.
	GeoIPCountryPath string `json:"geoip_country_path,omitempty"`

	// Hostname is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	Hostname *string `json:"hostname,omitempty"`

	// IgnoreBouncerError is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	IgnoreBouncerError *bool `json:"ignore_bouncer_error,omitempty"`

	// IgnoreOpenReportError is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	IgnoreOpenReportError *bool `json:"ignore_open_report_error,omitempty"`

	// MaxRuntime is the maximum runtime expressed. A negative
	// value for this field disables the maximum runtime. Using
	// a zero value will also mean disabled. This is not the
	// original behaviour of Measurement Kit, which used to run
	// for zero time in such case.
	MaxRuntime float64 `json:"max_runtime,omitempty"`

	// MLabNSAddressFamily is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	MLabNSAddressFamily *string `json:"mlabns/address_family,omitempty"`

	// MLabNSBaseURL is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	MLabNSBaseURL *string `json:"mlabns/base_url,omitempty"`

	// MLabNSCountry is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	MLabNSCountry *string `json:"mlabns/country,omitempty"`

	// MLabNSMetro is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	MLabNSMetro *string `json:"mlabns/metro,omitempty"`

	// MLabNSPolicy is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	MLabNSPolicy *string `json:"mlabns/policy,omitempty"`

	// MLabNSToolName is a legacy option that this library does
	// not support. Setting it causes the experiment to fail.
	MLabNSToolName *string `json:"mlabns_tool_name,omitempty"`

	// NoBouncer indicates whether to use a bouncer
	NoBouncer bool `json:"no_bouncer,omitempty"`

	// NoCollector indicates whether to use a collector
	NoCollector bool `json:"no_collector,omitempty"`

	// NoFileReport indicates whether to write a report file. Saving
	// the report to file is currently not implemented by this
	// library. Hence, if NoFileReport is false and OutputFilepath
	// is not empty, there will be a startup error.
	NoFileReport bool `json:"no_file_report,omitempty"`

	// NoGeoIP indicates whether to perform a GeoIP lookup. This
	// library fails if NoGeoIP and NoResolverLookup have different
	// values since these two steps are performed together.
	NoGeoIP bool `json:"no_geoip,omitempty"`

	// NoResolverLookup indicates whether to perform a resolver lookup. This
	// library fails if NoGeoIP and NoResolverLookup have different
	// values since these two steps are performed together.
	NoResolverLookup bool `json:"no_resolver_lookup"`

	// Port is the port used by performance tests. This library does not
	// support this option and fails if it is set by the user.
	Port *int64 `json:"port"`

	// ProbeASN is the AS number. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup warning, and the
	// library will otherwise ignore this setting.
	ProbeASN string `json:"probe_asn,omitempty"`

	// ProbeCC is the probe country code. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup warning, and the
	// library will otherwise ignore this setting.
	ProbeCC string `json:"probe_cc,omitempty"`

	// ProbeIP is the probe IP. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup warning, and the
	// library will otherwise ignore this setting.
	ProbeIP string `json:"probe_ip,omitempty"`

	// ProbeNetworkName is the probe network name. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup warning, and the
	// library will otherwise ignore this setting.
	ProbeNetworkName string `json:"probe_network_name,omitempty"`

	// RandomizeInput indicates whether to randomize inputs. This
	// option is not implemented by this library. Attempting
	// to set it to true will cause a startup error.
	RandomizeInput bool `json:"randomize_input,omitempty"`

	// SaveRealProbeIP indicates whether to save the real probe ASN
	SaveRealProbeASN bool `json:"save_real_probe_asn,omitempty"`

	// SaveRealProbeCC indicates whether to save the real probe CC
	SaveRealProbeCC bool `json:"save_real_probe_cc,omitempty"`

	// SaveRealProbeIP indicates whether to save the real probe IP
	SaveRealProbeIP bool `json:"save_real_probe_ip,omitempty"`

	// SaveRealResolverIP is a legacy option that this library
	// does not support. We will stop if you provide it.
	SaveRealResolverIP *bool `json:"save_real_resolver_ip,omitempty"`

	// Server is used by performance tests to indicate the specific
	// hostname that shall be used for the server. This library does
	// not support this setting and fails if you provide it.
	Server *string `json:"server,omitempty"`

	// SoftwareName is the software name. If this option is not
	// present, then the library startup will fail.
	SoftwareName string `json:"software_name,omitempty"`

	// SoftwareVersion is the software version. If this option is not
	// present, then the library startup will fail.
	SoftwareVersion string `json:"software_version,omitempty"`

	// TestSuite is a legacy option that this library does not support.
	TestSuite *int64 `json:"test_suite,omitempty"`

	// Timeout is a legacy option that this library does not support.
	Timeout *float64 `json:"timeout,omitempty"`

	// UUID is a legacy option that this library does not support.
	UUID *string `json:"uuid,omitempty"`
}
