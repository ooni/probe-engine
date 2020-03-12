// Package handlers contains default modelx.Handler handlers.
package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/netx/modelx"
)

type stdoutHandler struct{}

func (stdoutHandler) OnMeasurement(m modelx.Measurement) {
	data, err := json.Marshal(m)
	runtimex.PanicOnError(err, "unexpected json.Marshal failure")
	fmt.Printf("%s\n", string(data))
}

// StdoutHandler is a Handler that logs on stdout.
var StdoutHandler stdoutHandler

type noHandler struct{}

func (noHandler) OnMeasurement(m modelx.Measurement) {
}

// NoHandler is a Handler that does not print anything
var NoHandler noHandler
