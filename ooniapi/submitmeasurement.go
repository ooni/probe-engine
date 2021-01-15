package ooniapi

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

// SubmitMeasurementRequest is the SubmitMeasurement request.
type SubmitMeasurementRequest struct {
	ReportID string             `urlpath:"report_id"`
	Format   string             `json:"format"`
	Content  *model.Measurement `json:"content"`
	requestType
}

// SubmitMeasurementResponse is the SubmitMeasurement response.
type SubmitMeasurementResponse struct {
	responseType
}

// SubmitMeasurement implements the SubmitMeasurement API.
func (c Client) SubmitMeasurement(ctx context.Context, in *SubmitMeasurementRequest) (*SubmitMeasurementResponse, error) {
	var out SubmitMeasurementResponse
	err := c.api(ctx, apispec{
		Method:  "POST",
		URLPath: "/report/{report_id}",
		In:      in,
		Out:     &out,
	})
	return &out, err
}
