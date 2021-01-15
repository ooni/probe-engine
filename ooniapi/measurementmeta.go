package ooniapi

import "context"

// MeasurementMetaRequest is the MeasurementMeta Request.
type MeasurementMetaRequest struct {
	ReportID string `query:"report_id"`
	Full     bool   `query:"full"`
	Input    string `query:"input"`
	requestType
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
	responseType
}

// MeasurementMeta implements the MeasurementMeta API.
func (c Client) MeasurementMeta(ctx context.Context, in *MeasurementMetaRequest) (*MeasurementMetaResponse, error) {
	var out MeasurementMetaResponse
	err := c.api(ctx, apispec{
		Method:  "GET",
		URLPath: "/api/v1/measurement_meta",
		In:      in,
		Out:     &out,
	})
	return &out, err
}
