package handler_test

import (
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/handler"
)

func TestIntegration(t *testing.T) {
	printer := handler.NewPrinterCallbacks(log.Log)
	printer.OnDataUsage(10, 10)
	printer.OnProgress(0.4, "progress")
}
