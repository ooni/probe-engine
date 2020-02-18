package oonimkall

import (
	"sync/atomic"
	"time"
)

// TODO(bassosimone): event correctness wrt fields

// eventEmitter emits event on a channel
type eventEmitter struct {
	disabled map[string]bool
	out      chan<- *eventRecord
	timeouts int64
}

// newEventEmitter creates a new Emitter
func newEventEmitter(disabledEvents []string, out chan<- *eventRecord) *eventEmitter {
	ee := &eventEmitter{out: out}
	ee.disabled = make(map[string]bool)
	for _, eventname := range disabledEvents {
		ee.disabled[eventname] = true
	}
	return ee
}

// EmitFailureStartup emits the failureStartup event
func (ee *eventEmitter) EmitFailureStartup(failure string) {
	ee.EmitFailure(failureStartup, failure)
}

// EmitFailure emits a failure event
func (ee *eventEmitter) EmitFailure(name, failure string) {
	ee.Emit(name, eventValue{Failure: failure})
}

// EmitStatusProgress emits the status.Progress event
func (ee *eventEmitter) EmitStatusProgress(percentage float64, message string) {
	ee.Emit(statusProgress, eventValue{Message: message, Percentage: percentage})
}

// Emit emits the specified event
func (ee *eventEmitter) Emit(key string, value eventValue) {
	if ee.disabled[key] == true {
		return
	}
	select {
	case <-time.After(250 * time.Millisecond):
		atomic.AddInt64(&ee.timeouts, 1)
	case ee.out <- &eventRecord{Key: key, Value: value}:
	}
}
