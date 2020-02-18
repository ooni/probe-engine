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
	// to set it will cause a startup error.
	OutputFilePath string `json:"output_filepath,omitempty"`

	// StateDir is the directory where to store persistent data. This
	// field is an extension of MK's specification. If
	// this field is empty, the task won't start.
	StateDir string `json:"state_dir"`

	// TempDir is the temporary directory. This
	// field is an extension of MK's specification. If
	// this field is empty, the task won't start.
	TempDir string `json:"temp_dir"`
}

// TODO(bassosimone): restructure to have single "home" directory?

// settingsOptions contains the settings options
type settingsOptions struct {
	// Backend is a test helper for a nettest. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	Backend string `json:"backend,omitempty"`

	// BouncerBaseURL contains the bouncer base URL
	BouncerBaseURL string `json:"bouncer_base_url,omitempty"`

	// CaBundlePath contains the CA bundle path. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	CaBundlePath string `json:"net/ca_bundle_path,omitempty"`

	// CollectorBaseURL contains the collector base URL
	CollectorBaseURL string `json:"collector_base_url,omitempty"`

	// GeoIPASNPath is the ASN database path. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	GeoIPASNPath string `json:"geoip_asn_path,omitempty"`

	// GeoIPCountryPath is the country database path. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	GeoIPCountryPath string `json:"geoip_country_path,omitempty"`

	// MaxRuntime is the maximum runtime expressed. A negative
	// value for this filed disables the maximum runtime. Using
	// a zero value will cause the task to fail quickly.
	MaxRuntime float32 `json:"max_runtime,omitempty"`

	// NoBouncer indicates whether to use a bouncer
	NoBouncer bool `json:"no_bouncer,omitempty"`

	// NoCollector indicates whether to use a collector
	NoCollector bool `json:"no_collector,omitempty"`

	// NoFileReport indicates whether to write a report file. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	NoFileReport *bool `json:"no_file_report,omitempty"`

	// NoGeoIP indicates whether to perform a GeoIP lookup. This
	// library fails if NoGeoIP and NoResolverLookup have different
	// values since these two steps are performed together.
	NoGeoIP bool `json:"no_geoip,omitempty"`

	// NoResolverLookup indicates whether to perform a resolver lookup. This
	// library fails if NoGeoIP and NoResolverLookup have different
	// values since these two steps are performed together.
	NoResolverLookup bool `json:"no_resolver_lookup"`

	// ProbeASN is the AS number. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	ProbeASN string `json:"probe_asn,omitempty"`

	// ProbeCC is the probe country code. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	ProbeCC string `json:"probe_cc,omitempty"`

	// ProbeIP is the probe IP. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	ProbeIP string `json:"probe_ip,omitempty"`

	// ProbeNetworkName is the probe network name. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	ProbeNetworkName string `json:"probe_network_name,omitempty"`

	// RandomizeInput indicates whether to randomize inputs. This
	// option is not implemented by this library. Attempting
	// to set it will cause a startup error.
	RandomizeInput *bool `json:"randomize_input,omitempty"`

	// SaveRealProbeIP indicates whether to save the real probe IP
	SaveRealProbeIP bool `json:"save_real_probe_ip,omitempty"`

	// SaveRealProbeIP indicates whether to save the real probe ASN
	SaveRealProbeASN bool `json:"save_real_probe_asn,omitempty"`

	// SaveRealProbeCC indicates whether to save the real probe CC
	SaveRealProbeCC bool `json:"save_real_probe_cc,omitempty"`

	// SoftwareName is the software name. If this option is not
	// present, then the library startup will fail.
	SoftwareName string `json:"software_name,omitempty"`

	// SoftwareVersion is the software version. If this option is not
	// present, then the library startup will fail.
	SoftwareVersion string `json:"software_version,omitempty"`
}
