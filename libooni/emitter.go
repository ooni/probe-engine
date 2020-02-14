package main

import (
	"time"
)

// Emitter emits event on a channel
type Emitter struct {
	disabled map[string]bool
	out      chan<- *Event
}

// NewEmitter creates a new Emitter
func NewEmitter(
	settings Settings,
	out chan<- *Event,
) *Emitter {
	ee := &Emitter{out: out}
	for _, eventname := range settings.DisabledEvents {
		ee.disabled[eventname] = true
	}
	return ee
}

// EmitFailureStartup emits the failureStartup event
func (ee *Emitter) EmitFailureStartup(failure string) {
	ee.EmitFailure(failureStartup, failure)
}

// EmitFailure emits a failure event
func (ee *Emitter) EmitFailure(name, failure string) {
	ee.Emit(name, EventValue{Failure: failure})
}

// EmitStatusProgress emits the status.Progress event
func (ee *Emitter) EmitStatusProgress(percentage float64, message string) {
	ee.Emit(statusProgress, EventValue{Message: message, Percentage: percentage})
}

// Emit emits the specified event
func (ee *Emitter) Emit(key string, value EventValue) {
	if ee.disabled[key] == true {
		return
	}
	select {
	case <-time.After(250 * time.Millisecond):
	case ee.out <- &Event{Key: key, Value: value}:
	}
}
