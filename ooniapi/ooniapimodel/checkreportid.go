package ooniapimodel

// CheckReportIDRequest is the CheckReportID request.
type CheckReportIDRequest struct {
	ReportID string `query:"report_id"`
	RequestType
}

// CheckReportIDResponse is the CheckReportID response.
type CheckReportIDResponse struct {
	Found bool `json:"found"`
	ResponseType
}

// GETCheckReportID is the GET /api/_/check_report_id API call.
type GETCheckReportID struct {
	Method   MethodType  `method:"GET"`
	URLPath  URLPathType `path:"/api/_/check_report_id"`
	Request  CheckReportIDRequest
	Response CheckReportIDResponse
	APIType
}
