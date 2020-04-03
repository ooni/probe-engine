package resolver

import (
	"sync"

	"github.com/ooni/probe-engine/netx/modelx"
)

type SavingHandler struct {
	mu sync.Mutex
	v  []modelx.Measurement
}

func (sh *SavingHandler) OnMeasurement(ev modelx.Measurement) {
	sh.mu.Lock()
	sh.v = append(sh.v, ev)
	sh.mu.Unlock()
}

func (sh *SavingHandler) Read() []modelx.Measurement {
	sh.mu.Lock()
	v := sh.v
	sh.v = nil
	sh.mu.Unlock()
	return v
}
