// Package mkevent processes MK events
package mkevent

import (
	"encoding/json"
	"strings"

	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/measurementkit"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

// Handle handles a measurementkit event
func Handle(
	sess *session.Session,
	measurement *model.Measurement,
	event measurementkit.Event,
	callbacks handler.Callbacks,
) {
	if event.Key == "measurement" {
		// We reparse the measurement and overwrite it. This is how we manage to
		// return the measurement to the caller. Seriousy.
		//
		// We panic if we cannot parse because since MK v0.9.0 the vendored
		// nlohmann/json library should throw if passed non UTF-8 input.
		err := json.Unmarshal([]byte(event.Value.JSONStr), measurement)
		if err != nil {
			panic(err)
		}
		return
	}
	if event.Key == "log" {
		if strings.HasPrefix(event.Value.LogLevel, "DEBUG") {
			sess.Logger.Debug(event.Value.Message)
		} else if event.Value.LogLevel == "INFO" {
			sess.Logger.Info(event.Value.Message)
		} else {
			sess.Logger.Warn(event.Value.Message)
		}
		return
	}
	if event.Key == "status.progress" {
		callbacks.OnProgress(event.Value.Percentage, event.Value.Message)
		return
	}
	if event.Key == "status.end" {
		callbacks.OnDataUsage(event.Value.DownloadedKB, event.Value.UploadedKB)
		return
	}
	/*
		if event.Key == "status.update.performance" {
			return // Seems unused by OONI
		}
		if event.Key == "status.update.websites" {
			return // Ditto
		}
		if event.Key == "status.queued" {
			return // Ditto
		}
		if event.Key == "status.started" {
			return // Ditto
		}
		if event.Key == "status.measurement_start" {
			return // We know that because we control the lifecycle
		}
		if event.Key == "status.measurement_done" {
			return // We know that because we control the lifecycle
		}
		if event.Key == "status.geoip_lookup" {
			return // We perform the lookup before calling MK
		}
	*/
	sess.Logger.Debugf("mkevent: %s %+v", event.Key, event.Value)
}
