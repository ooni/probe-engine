// Package netxlogger contains a simple logger for netx
package netxlogger

import (
	"encoding/json"

	"github.com/ooni/netx/modelx"
)

// Logger is the logger this package expects
type Logger interface {
	Debug(msg string)
}

// Handler is a measurements handler
type Handler struct {
	logger Logger
}

// New creates a new log handler
func New(logger Logger) *Handler {
	return &Handler{logger: logger}
}

// OnMeasurement handles the measurement
func (h *Handler) OnMeasurement(m modelx.Measurement) {
	if data, err := json.Marshal(m); err == nil {
		h.logger.Debug(string(data))
	}
}
