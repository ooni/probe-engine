package ooniapimodel

// MeasurementMetaRequest is the MeasurementMeta Request.
type MeasurementMetaRequest struct {
	ReportID string `query:"report_id"`
	Full     bool   `query:"full"`
	Input    string `query:"input"`
	RequestType
}

// MeasurementMetaResponse is the MeasurementMeta Response.
type MeasurementMetaResponse struct {
	Anomaly              bool   `json:"anomaly"`
	CategoryCode         string `json:"category_code"`
	Confirmed            bool   `json:"confirmed"`
	Failure              bool   `json:"failure"`
	Input                string `json:"input"`
	MeasurementStartTime string `json:"measurement_start_time"`
	ProbeASN             int64  `json:"probe_asn"`
	ProbeCC              string `json:"probe_cc"`
	RawMeasurement       string `json:"raw_measurement"`
	ReportID             string `json:"report_id"`
	Scores               string `json:"scores"`
	TestName             string `json:"test_name"`
	TestStartTime        string `json:"test_start_time"`
	ResponseType
}

// GETMeasurementMeta is the GET /api/v1/measurement_meta API call.
type GETMeasurementMeta struct {
	Method   MethodType  `method:"GET"`
	URLPath  URLPathType `path:"/api/v1/measurement_meta"`
	Request  MeasurementMetaRequest
	Response MeasurementMetaResponse
	APIType
}
