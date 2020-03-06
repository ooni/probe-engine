package mkevent_test

import (
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/mkevent"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
)

func TestIntegrationMeasurementSuccess(t *testing.T) {
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	var m model.Measurement
	printer := handler.NewPrinterCallbacks(log.Log)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key: "measurement",
		Value: measurementkit.EventValue{
			JSONStr: "{}",
		},
	}, printer)
}

func TestIntegrationMeasurementFailure(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic here")
		}
	}()
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	var m model.Measurement
	printer := handler.NewPrinterCallbacks(log.Log)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key: "measurement",
		Value: measurementkit.EventValue{
			JSONStr: "{",
		},
	}, printer)
}

func TestIntegrationLog(t *testing.T) {
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	var m model.Measurement
	printer := handler.NewPrinterCallbacks(log.Log)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key: "log",
		Value: measurementkit.EventValue{
			LogLevel: "DEBUG",
			Message:  "message",
		},
	}, printer)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key: "log",
		Value: measurementkit.EventValue{
			LogLevel: "INFO",
			Message:  "message",
		},
	}, printer)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key: "log",
		Value: measurementkit.EventValue{
			LogLevel: "WARNING",
			Message:  "message",
		},
	}, printer)
}

func TestIntegrationStatusProgress(t *testing.T) {
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	var m model.Measurement
	printer := handler.NewPrinterCallbacks(log.Log)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key: "status.progress",
		Value: measurementkit.EventValue{
			Percentage: 0.17,
			Message:    "message",
		},
	}, printer)
}

func TestIntegrationStatusEnd(t *testing.T) {
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	var m model.Measurement
	printer := handler.NewPrinterCallbacks(log.Log)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key: "status.end",
		Value: measurementkit.EventValue{
			DownloadedKB: 1234,
			UploadedKB:   5678,
		},
	}, printer)
}

func TestIntegrationOtherEvent(t *testing.T) {
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	var m model.Measurement
	printer := handler.NewPrinterCallbacks(log.Log)
	mkevent.Handle(sess, &m, measurementkit.Event{
		Key:   "status.queued",
		Value: measurementkit.EventValue{},
	}, printer)
}
