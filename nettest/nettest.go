// Package nettest contains generic nettest code.
package nettest

import (
	"context"
	"errors"
	"time"

	"github.com/ooni/probe-engine/collector"
	"github.com/ooni/probe-engine/geoiplookup/constants"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

// dateFormat is the format used by OONI for dates inside reports.
const dateFormat = "2006-01-02 15:04:05"

// formatTimeNowUTC formats the current time in UTC using the OONI format.
func formatTimeNowUTC() string {
	return time.Now().UTC().Format(dateFormat)
}

// Nettest is a generic nettest.
type Nettest struct {
	// Report is the report bound to this nettest.
	Report *collector.Report

	// Session is the session to which this nettest belongs.
	Session *session.Session

	// TestName is the nettest name.
	TestName string

	// TestStartTime is the UTC time when the test started.
	TestStartTime string

	// TestVersion is the nettest version.
	TestVersion string
}

// New creates a new nettest.
func New(session *session.Session, testName, testVersion string) *Nettest {
	return &Nettest{
		Session:       session,
		TestName:      testName,
		TestStartTime: formatTimeNowUTC(),
		TestVersion:   testVersion,
	}
}

// OpenReport opens a new report for the nettest.
func (n *Nettest) OpenReport(ctx context.Context) (err error) {
	if n.Report != nil {
		return // already open
	}
	for _, e := range n.Session.AvailableCollectors {
		if e.Type != "https" {
			n.Session.Logger.Debugf("nettest: unsupported collector type: %s", e.Type)
			continue
		}
		client := &collector.Client{
			BaseURL:    e.Address,
			HTTPClient: n.Session.HTTPDefaultClient, // proxy is OK
			Logger:     n.Session.Logger,
			UserAgent:  n.Session.UserAgent(),
		}
		template := collector.ReportTemplate{
			ProbeASN:        n.Session.ProbeASN,
			ProbeCC:         n.Session.ProbeCC,
			SoftwareName:    n.Session.SoftwareName,
			SoftwareVersion: n.Session.SoftwareVersion,
			TestName:        n.TestName,
			TestVersion:     n.TestVersion,
		}
		n.Report, err = client.OpenReport(ctx, template)
		if err == nil {
			return
		}
		n.Session.Logger.Debugf("nettest: collector error: %s", err.Error())
	}
	err = errors.New("All collectors failed")
	return
}

func (n *Nettest) reportID() string {
	if n.Report == nil {
		return ""
	}
	return n.Report.ID
}

// NewMeasurement initializes and returns a new measurement. Note that this
// function will set the ProbeIP to the default, privacy preserving value. If
// you need the ProbeIP when running a nettest and/or you want to submit it
// because the user asked for that, please override its value explicitly.
func (n *Nettest) NewMeasurement(input string) model.Measurement {
	return model.Measurement{
		DataFormatVersion:    "0.2.0",
		Input:                input,
		MeasurementStartTime: formatTimeNowUTC(),
		ProbeIP:              constants.DefaultProbeIP, // privacy by default
		ProbeASN:             n.Session.ProbeASN,
		ProbeCC:              n.Session.ProbeCC,
		ReportID:             n.reportID(),
		SoftwareName:         n.Session.SoftwareName,
		SoftwareVersion:      n.Session.SoftwareVersion,
		TestName:             n.TestName,
		TestStartTime:        n.TestStartTime,
		TestVersion:          n.TestVersion,
	}
}

// SubmitMeasurement submits a measurement to the selected collector. It is
// safe to call this function from different goroutines concurrently as long
// as the measurement is not shared by the goroutines.
func (n *Nettest) SubmitMeasurement(
	ctx context.Context, measurement *model.Measurement,
) (err error) {
	if n.Report != nil {
		err = n.Report.SubmitMeasurement(ctx, measurement)
	}
	return
}

// CloseReport closes the open report.
func (n *Nettest) CloseReport(ctx context.Context) (err error) {
	if n.Report != nil {
		err = n.Report.Close(ctx)
		n.Report = nil
	}
	return
}
