// Package handler contains experiment events handler
package handler

import (
	"github.com/ooni/probe-engine/log"
)

// Callbacks contains event handling callbacks
type Callbacks interface {
	// DataUsage provides information about data usage.
	DataUsage(dloadKiB, uploadKiB float64)

	// Progress provides information about an experiment progress.
	Progress(percentage float64, message string)
}

// PrinterCallbacks is the default event handler
type PrinterCallbacks struct {
	log.Logger
}

// NewPrinterCallbacks returns a new default callback handler
func NewPrinterCallbacks(logger log.Logger) PrinterCallbacks {
	return PrinterCallbacks{Logger: logger}
}

// DataUsage provides information about data usage.
func (d PrinterCallbacks) DataUsage(dloadKiB, uploadKiB float64) {
	d.Logger.Infof("data usage: %f/%f down/up KiB", dloadKiB, uploadKiB)
}

// Progress provides information about an experiment progress.
func (d PrinterCallbacks) Progress(percentage float64, message string) {
	d.Logger.Infof("[%4.1f%%] %s", percentage*100, message)
}
