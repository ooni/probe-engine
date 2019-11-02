package netxlogger_test

import (
	"testing"

	"github.com/ooni/netx/modelx"
	"github.com/ooni/probe-engine/experiment/netxlogger"
)

func TestIntegration(t *testing.T) {
	logger := &fakelogger{}
	handler := netxlogger.New(logger)
	handler.OnMeasurement(modelx.Measurement{})
	if logger.called == false {
		t.Fatal("not called")
	}
}

type fakelogger struct {
	called bool
}

func (f *fakelogger) Debug(msg string) {
	f.called = true
}
