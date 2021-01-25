package ooniapimodel

import "github.com/ooni/probe-engine/model"

// SubmitMeasurementRequest is the SubmitMeasurement request.
type SubmitMeasurementRequest struct {
	ReportID string             `urlpath:"report_id"`
	Format   string             `json:"format"`
	Content  *model.Measurement `json:"content"`
	RequestType
}

// SubmitMeasurementResponse is the SubmitMeasurement response.
type SubmitMeasurementResponse struct {
	ResponseType
}

// POSTSubmitMeasurement is the POST /measurement/{report_id} API call.
type POSTSubmitMeasurement struct {
	Method   MethodType  `method:"POST"`
	URLPath  URLPathType `path:"/report/{report_id}"`
	Request  SubmitMeasurementRequest
	Response SubmitMeasurementResponse
	APIType
}
