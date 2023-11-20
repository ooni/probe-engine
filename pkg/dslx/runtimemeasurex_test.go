package dslx

import (
	"testing"
	"time"

	"github.com/ooni/probe-engine/pkg/measurexlite"
	"github.com/ooni/probe-engine/pkg/mocks"
	"github.com/ooni/probe-engine/pkg/model"
)

func TestMeasurexLiteRuntime(t *testing.T) {
	t.Run("we can configure a custom model.MeasuringNetwork", func(t *testing.T) {
		netx := &mocks.MeasuringNetwork{}
		rt := NewRuntimeMeasurexLite(model.DiscardLogger, time.Now(), RuntimeMeasurexLiteOptionMeasuringNetwork(netx))
		if rt.netx != netx {
			t.Fatal("did not set the measuring network")
		}
		trace := rt.NewTrace(rt.IDGenerator().Add(1), rt.ZeroTime()).(*measurexlite.Trace)
		if trace.Netx != netx {
			t.Fatal("did not set the measuring network")
		}
	})
}
