package tasks

// eventEmitter emits event on a channel
type eventEmitter struct {
	disabled map[string]bool
	out      chan<- *Event
}

// newEventEmitter creates a new Emitter
func newEventEmitter(disabledEvents []string, out chan<- *Event) *eventEmitter {
	ee := &eventEmitter{out: out}
	ee.disabled = make(map[string]bool)
	for _, eventname := range disabledEvents {
		ee.disabled[eventname] = true
	}
	return ee
}

// EmitFailureStartup emits the failureStartup event
func (ee *eventEmitter) EmitFailureStartup(failure string) {
	ee.EmitFailureGeneric(failureStartup, failure)
}

// EmitFailure emits a failure event
func (ee *eventEmitter) EmitFailureGeneric(name, failure string) {
	ee.Emit(name, eventFailureGeneric{Failure: failure})
}

// EmitStatusProgress emits the status.Progress event
func (ee *eventEmitter) EmitStatusProgress(percentage float64, message string) {
	ee.Emit(statusProgress, eventStatusProgress{Message: message, Percentage: percentage})
}

// Emit emits the specified event
func (ee *eventEmitter) Emit(key string, value interface{}) {
	if ee.disabled[key] == true {
		return
	}
	ee.out <- &Event{Key: key, Value: value}
}
