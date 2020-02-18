package oonimkall

// eventValue are all the possible value keys
type eventValue struct {
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

	// ResolverASN is the resolver ASN
	ResolverASN string `json:"resolver_asn,omitempty"`

	// ResolverIP is the resolver IP
	ResolverIP string `json:"resolver_ip,omitempty"`

	// ResolverNetworkName is the resolver network name
	ResolverNetworkName string `json:"resolver_network_name,omitempty"`

	// UploadedKB is the amount of uploaded KiBs
	UploadedKB float64 `json:"uploaded_kb,omitempty"`
}

// eventRecord is an event emitted by a task. This structure extends the event
// described by MK v0.10.9 FFI API (https://git.io/Jv4Rv).
type eventRecord struct {
	// Is the key for the event
	Key string `json:"key"`

	// Contains the value for the event
	Value eventValue `json:"value"`
}
