package model_test

import (
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/model"
)

func TestPrinterCallbacksCallbacks(t *testing.T) {
	printer := model.NewPrinterCallbacks(log.Log)
	printer.OnProgress(0.4, "progress")
}
