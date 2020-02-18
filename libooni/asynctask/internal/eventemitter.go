package internal

import (
	"time"
)

// TODO(bassosimone): event correctness wrt fields

// eventEmitter emits event on a channel
type eventEmitter struct {
	disabled map[string]bool
	out      chan<- *Event
}

// newEventEmitter creates a new Emitter
func newEventEmitter(
	settings Settings,
	out chan<- *Event,
) *eventEmitter {
	ee := &eventEmitter{out: out}
	for _, eventname := range settings.DisabledEvents {
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
	ee.Emit(name, EventValue{Failure: failure})
}

// EmitStatusProgress emits the status.Progress event
func (ee *eventEmitter) EmitStatusProgress(percentage float64, message string) {
	ee.Emit(statusProgress, EventValue{Message: message, Percentage: percentage})
}

// Emit emits the specified event
func (ee *eventEmitter) Emit(key string, value EventValue) {
	if ee.disabled[key] == true {
		return
	}
	select {
	case <-time.After(250 * time.Millisecond):
	case ee.out <- &Event{Key: key, Value: value}:
	}
}
