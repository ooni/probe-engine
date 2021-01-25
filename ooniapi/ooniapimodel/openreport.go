package ooniapimodel

// OpenReportRequest is the OpenReport request.
type OpenReportRequest struct {
	DataFormatVersion string `json:"data_format_version"`
	Format            string `json:"format"`
	ProbeASN          string `json:"probe_asn"`
	ProbeCC           string `json:"probe_cc"`
	SoftwareName      string `json:"software_name"`
	SoftwareVersion   string `json:"software_version"`
	TestName          string `json:"test_name"`
	TestStartTime     string `json:"test_start_time"`
	TestVersion       string `json:"test_version"`
	RequestType
}

// OpenReportResponse is the OpenReport response.
type OpenReportResponse struct {
	ReportID         string   `json:"report_id"`
	SupportedFormats []string `json:"supported_formats"`
	ResponseType
}

// POSTOpenReport is the POST /report API call.
type POSTOpenReport struct {
	Method   MethodType  `method:"POST"`
	URLPath  URLPathType `path:"/report"`
	Request  OpenReportRequest
	Response OpenReportResponse
	APIType
}
