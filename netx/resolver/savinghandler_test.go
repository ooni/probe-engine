package resolver

import (
	"sync"

	"github.com/ooni/probe-engine/netx/modelx"
)

// SavingHandler saves the events
type SavingHandler struct {
	mu sync.Mutex
	v  []modelx.Measurement
}

// OnMeasurement implements modelx.Handler.OnMeasurement
func (sh *SavingHandler) OnMeasurement(ev modelx.Measurement) {
	sh.mu.Lock()
	sh.v = append(sh.v, ev)
	sh.mu.Unlock()
}

// Read reads the measurements
func (sh *SavingHandler) Read() []modelx.Measurement {
	sh.mu.Lock()
	v := sh.v
	sh.v = nil
	sh.mu.Unlock()
	return v
}
