package ooniapi

import "context"

// CheckReportIDRequest is the CheckReportID request.
type CheckReportIDRequest struct {
	ReportID string `query:"report_id"`
	requestType
}

// CheckReportIDResponse is the CheckReportID response.
type CheckReportIDResponse struct {
	Found bool `json:"found"`
	responseType
}

// CheckReportID implements the CheckReportID API.
func (c Client) CheckReportID(ctx context.Context, in *CheckReportIDRequest) (*CheckReportIDResponse, error) {
	var out CheckReportIDResponse
	err := c.api(ctx, apispec{
		Method:  "GET",
		URLPath: "/api/_/check_report_id",
		In:      in,
		Out:     &out,
	})
	return &out, err
}
