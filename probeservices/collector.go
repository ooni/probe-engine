package probeservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ooni/probe-engine/model"
)

const (
	// DefaultDataFormatVersion is the default data format version.
	//
	// See https://github.com/ooni/spec/tree/master/data-formats#history.
	DefaultDataFormatVersion = "0.2.0"

	// DefaultFormat is the default format
	DefaultFormat = "json"
)

// ReportTemplate is the template for opening a report
type ReportTemplate struct {
	// DataFormatVersion is unconditionally set to DefaultDataFormatVersion
	// and you don't need to be concerned about it.
	DataFormatVersion string `json:"data_format_version"`

	// Format is unconditionally set to `json` and you don't need
	// to be concerned about it.
	Format string `json:"format"`

	// ProbeASN is the probe's autonomous system number (e.g. `AS1234`)
	ProbeASN string `json:"probe_asn"`

	// ProbeCC is the probe's country code (e.g. `IT`)
	ProbeCC string `json:"probe_cc"`

	// SoftwareName is the app name (e.g. `measurement-kit`)
	SoftwareName string `json:"software_name"`

	// SoftwareVersion is the app version (e.g. `0.9.1`)
	SoftwareVersion string `json:"software_version"`

	// TestName is the test name (e.g. `ndt`)
	TestName string `json:"test_name"`

	// TestVersion is the test version (e.g. `1.0.1`)
	TestVersion string `json:"test_version"`
}

type openResponse struct {
	ID               string   `json:"report_id"`
	SupportedFormats []string `json:"supported_formats"`
}

// Report is an open report
type Report struct {
	// ID is the report ID
	ID string

	// client is the client that was used.
	client Client
}

// OpenReport opens a new report.
func (c Client) OpenReport(ctx context.Context, rt ReportTemplate) (*Report, error) {
	if rt.DataFormatVersion != DefaultDataFormatVersion {
		return nil, errors.New("Unsupported data format version")
	}
	if rt.Format != DefaultFormat {
		return nil, errors.New("Unsupported format")
	}
	var or openResponse
	err := c.Client.CreateJSON(ctx, "/report", rt, &or)
	if err != nil {
		return nil, err
	}
	for _, format := range or.SupportedFormats {
		if format == "json" {
			return &Report{ID: or.ID, client: c}, nil
		}
	}
	return nil, errors.New("JSON format not supported")
}

type updateRequest struct {
	// Format is the data format
	Format string `json:"format"`

	// Content is the actual report
	Content interface{} `json:"content"`
}

type updateResponse struct {
	// ID is the measurement ID
	ID string `json:"measurement_id"`
}

// SubmitMeasurement submits a measurement belonging to the report
// to the OONI collector. We will unconditionally modify the measurement
// with the ReportID it should contain. If the collector supports sending
// back to us a measurement ID, we also update the m.OOID field with it.
func (r Report) SubmitMeasurement(ctx context.Context, m *model.Measurement) error {
	var updateResponse updateResponse
	m.ReportID = r.ID
	err := r.client.Client.CreateJSON(
		ctx, fmt.Sprintf("/report/%s", r.ID), updateRequest{
			Format:  "json",
			Content: m,
		}, &updateResponse,
	)
	if err == nil {
		m.OOID = updateResponse.ID
	}
	return err
}

// Close closes the report. Returns nil on success; an error on failure.
func (r Report) Close(ctx context.Context) error {
	var input, output struct{}
	err := r.client.Client.CreateJSON(
		ctx, fmt.Sprintf("/report/%s/close", r.ID), input, &output,
	)
	// Implementation note: the server is not compliant with
	// the spec, which says it MUST return a JSON. It does
	// instead return an empty string. Intercept this error
	// and turn it to nil, since we cannot really act upon
	// this error, and we ought be flexible.
	if _, ok := err.(*json.SyntaxError); ok && err.Error() == "unexpected end of JSON input" {
		r.client.Logger.Debug(
			"collector.go: working around collector returning empty string bug",
		)
		err = nil
	}
	return err
}
