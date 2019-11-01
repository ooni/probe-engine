// Package netxlogger contains a simple logger for netx
package netxlogger

import (
	"encoding/json"

	"github.com/ooni/probe-engine/log"
	"github.com/ooni/netx/model"
)

// Handler is a measurements handler
type Handler struct {
	logger log.Logger
}

// New creates a new log handler
func New(logger log.Logger) *Handler {
	return &Handler{logger: logger}
}

// OnMeasurement handles the measurement
func (h *Handler) OnMeasurement(m model.Measurement) {
	if data, err := json.Marshal(m); err == nil {
		h.logger.Debug(string(data))
	}
}
