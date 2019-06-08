package oonimobile

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"strings"
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

	// Logs contains the task logs. This field is only filled when you
	// run the task in a synchronous fashion. It is deprecated.
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
		SoftwareName:          softwareName,
		SoftwareVersion:       softwareVersion,
	}
}

// ResubmitHandle is an opaque reference to an async resubmission
// task that is running in a background thread.
type ResubmitHandle struct {
	Results *ResubmitResults
	cancel  context.CancelFunc
	ch      chan *LogMessage
	ctx     context.Context
	task    *ResubmitTask
}

func (rh *ResubmitHandle) do() {
	logger := &channelLogger{rh.ch}
	defer close(rh.ch)
	defer rh.cancel()
	var measurement model.Measurement
	err := json.Unmarshal([]byte(rh.task.SerializedMeasurement), &measurement)
	if err != nil {
		logger.Warnf("resubmit: cannot unmarshal JSON: %s", err.Error())
		return
	}
	// Note that time.Duration is int64
	const maxTimeout = math.MaxInt64 / int64(time.Second)
	if rh.task.Timeout <= 0 || rh.task.Timeout >= maxTimeout {
		rh.task.Timeout = maxTimeout
	}
	ctx, cancel := context.WithTimeout(
		rh.ctx, time.Duration(rh.task.Timeout)*time.Second,
	)
	defer cancel()
	sess := session.New(
		logger, rh.task.SoftwareName, rh.task.SoftwareVersion,
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
	rh.Results.UpdatedSerializedMeasurement = string(data)
	rh.Results.UpdatedReportID = measurement.ReportID
	rh.Results.Good = true
	return
}

// Interrupt interrupts the async resubmission task.
func (rh *ResubmitHandle) Interrupt() {
	rh.cancel()
}

// WaitForLogMessage blocks until the next log message is available, while
// the task is running, and returns it. If the task is not running, instead,
// it immediately returns nil.
func (rh *ResubmitHandle) WaitForLogMessage() *LogMessage {
	return <-rh.ch
}

// Start starts an async resubmit task.
func (rt *ResubmitTask) Start() *ResubmitHandle {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *LogMessage)
	rh := &ResubmitHandle{
		Results: new(ResubmitResults),
		cancel:  cancel,
		ch:      ch,
		ctx:     ctx,
		task:    rt,
	}
	go rh.do()
	return rh
}

// Run performs the resubmission task in a sync fashion.
//
// This method is deprecated and will be removed.
func (rt *ResubmitTask) Run() *ResubmitResults {
	rh := rt.Start()
	var builder strings.Builder
	for {
		logMessage := rh.WaitForLogMessage()
		if logMessage == nil {
			break
		}
		builder.WriteString("<" + logMessage.Level + "> ")
		builder.WriteString(logMessage.Message)
		builder.WriteString("\n")
	}
	rh.Results.Logs = builder.String()
	return rh.Results
}
