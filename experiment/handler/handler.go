// Package handler contains experiment events handler
package handler

import (
	"github.com/dustin/go-humanize"
	"github.com/ooni/probe-engine/model"
)

// PrinterCallbacks is the default event handler
type PrinterCallbacks struct {
	model.Logger
}

// NewPrinterCallbacks returns a new default callback handler
func NewPrinterCallbacks(logger model.Logger) PrinterCallbacks {
	return PrinterCallbacks{Logger: logger}
}

// OnDataUsage provides information about data usage.
func (d PrinterCallbacks) OnDataUsage(dloadKiB, uploadKiB float64) {
	d.Logger.Infof("experiment: recv %s, sent %s",
		humanize.SI(dloadKiB*1024, "byte"),
		humanize.SI(uploadKiB*1024, "byte"),
	)
}

// OnProgress provides information about an experiment progress.
func (d PrinterCallbacks) OnProgress(percentage float64, message string) {
	d.Logger.Infof("[%4.1f%%] %s", percentage*100, message)
}
