// Package collector contains a OONI collector client implementation.
//
// Specifically we implement v2.0.0 of the OONI collector specification defined
// in https://github.com/ooni/spec/blob/master/backends/bk-003-collector.md.
package collector

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ooni/probe-engine/httpx/jsonapi"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
)

// Client is a client for the OONI collector API.
type Client struct {
	// BaseURL is the bouncer base URL.
	BaseURL string

	// HTTPClient is the HTTP client to use.
	HTTPClient *http.Client

	// Logger is the logger to use.
	Logger log.Logger

	// UserAgent is the user agent to use.
	UserAgent string
}

// ReportTemplate is the template for opening a report
type ReportTemplate struct {
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

// Report is an open report
type Report struct {
	// ID is the report ID
	ID string `json:"report_id"`

	// Client is the client that was used.
	Client *Client
}

// OpenReport opens a new report.
func (c *Client) OpenReport(
	ctx context.Context, rt ReportTemplate,
) (*Report, error) {
	var report Report
	err := (&jsonapi.Client{
		BaseURL:    c.BaseURL,
		HTTPClient: c.HTTPClient,
		Logger:     c.Logger,
		UserAgent:  c.UserAgent,
	}).Create(ctx, "/report", rt, &report)
	report.Client = c
	return &report, err
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
// to the OONI collector. If the collector supports sending back to
// us a measurement ID, we also update the m.OOID field with it.
func (r *Report) SubmitMeasurement(
	ctx context.Context, m *model.Measurement, includeProbeIP bool,
) error {
	var updateResponse updateResponse
	if !includeProbeIP {
		m.ProbeIP = model.DefaultProbeIP
	}
	err := (&jsonapi.Client{
		BaseURL:    r.Client.BaseURL,
		HTTPClient: r.Client.HTTPClient,
		Logger:     r.Client.Logger,
		UserAgent:  r.Client.UserAgent,
	}).Create(
		ctx, fmt.Sprintf("/report/%s", r.ID), updateRequest{
			Format:  "json",
			Content: m,
		}, &updateResponse,
	)
	m.OOID = updateResponse.ID
	return err
}

// Close closes the report. Returns nil on success; an error on failure.
func (r *Report) Close(ctx context.Context) error {
	var input, output struct{}
	return (&jsonapi.Client{
		BaseURL:    r.Client.BaseURL,
		HTTPClient: r.Client.HTTPClient,
		Logger:     r.Client.Logger,
		UserAgent:  r.Client.UserAgent,
	}).Create(
		ctx, fmt.Sprintf("/report/%s/close", r.ID), input, &output,
	)
}
