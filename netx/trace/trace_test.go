package trace_test

import (
	"sync"
	"testing"

	"github.com/ooni/probe-engine/netx/trace"
)

func TestIntegration(t *testing.T) {
	saver := trace.Saver{}
	var wg sync.WaitGroup
	const parallel = 10
	wg.Add(parallel)
	for idx := 0; idx < parallel; idx++ {
		go func() {
			saver.Write(trace.Event{})
			wg.Done()
		}()
	}
	wg.Wait()
	ev := saver.Read()
	if len(ev) != parallel {
		t.Fatal("unexpected number of events read")
	}
}
