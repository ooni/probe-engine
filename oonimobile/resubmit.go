package oonimobile

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"time"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

// XXX: actually use CA bundle path

// ResubmitTask is a task for resubmitting measurements
type ResubmitTask struct {
	// CABundlePath is the CA bundle path to use
	CABundlePath string

	// SerializedMeasurement is the measurement to resubmit
	SerializedMeasurement string

	// SoftwareName is the app name
	SoftwareName string

	// SoftwareVersion is the app version
	SoftwareVersion string

	// Timeout is the whole-operation timeout in seconds
	Timeout int64
}

// ResubmitResults contains the results of resubmitting a report
type ResubmitResults struct {
	// Good indicates whether we succeeded
	Good bool

	// Logs contains the task logs
	Logs string

	// UpdatedReportID contains the updated report ID
	UpdatedReportID string

	// UpdatedSerializedMeasurement contains the updated serialized measurement
	UpdatedSerializedMeasurement string
}

// NewResubmitTask creates a new resubmit task. It is your responsibility
// to provide a measurement where all fields that should be redacted
// according to privacy settings have been redacted.
func NewResubmitTask(
	softwareName string,
	softwareVersion string,
	serializedMeasurement string,
) *ResubmitTask {
	return &ResubmitTask{
		SerializedMeasurement: serializedMeasurement,
		SoftwareName: softwareName,
		SoftwareVersion: softwareVersion,
	}
}

// Run performs the resubmission task.
func (rt *ResubmitTask) Run() (out *ResubmitResults) {
	out = new(ResubmitResults)
	logger := new(stringLogger)
	defer logger.SaveLogsInto(&out.Logs)
	var measurement model.Measurement
	err := json.Unmarshal([]byte(rt.SerializedMeasurement), &measurement)
	if err != nil {
		logger.Warnf("resubmit: cannot unmarshal JSON: %s", err.Error())
		return
	}
	// Note that time.Duration is int64
	const maxTimeout = math.MaxInt64 / int64(time.Second)
	if rt.Timeout <= 0 || rt.Timeout >= maxTimeout {
		rt.Timeout = maxTimeout
	}
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(rt.Timeout) * time.Second,
	)
	defer cancel()
	sess := session.New(
		logger, rt.SoftwareName, rt.SoftwareVersion,
		os.TempDir(), // use temporary directory to be safe
		nil, nil,
	)
	report, err := sess.OpenReport(
		ctx, measurement.TestName, measurement.TestVersion,
	)
	if err != nil {
		return
	}
	defer report.Close(ctx)
	measurement.ReportID = report.ID
	err = report.SubmitMeasurement(ctx, &measurement)
	if err != nil {
		return
	}
	data, err := json.Marshal(&measurement)
	if err != nil {
		logger.Warnf("resubmit: cannot marshal JSON: %s", err.Error())
		return
	}
	out.UpdatedSerializedMeasurement = string(data)
	out.UpdatedReportID = measurement.ReportID
	out.Good = true
	return
}
