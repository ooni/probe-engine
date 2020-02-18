package oonimkall

import (
	"sync"
	"testing"
)

func TestUnitDisabledEvents(t *testing.T) {
	out := make(chan *eventRecord)
	emitter := newEventEmitter([]string{"log"}, out)
	go func() {
		emitter.Emit("log", eventValue{Message: "foo"})
		close(out)
	}()
	var count int64
	for ev := range out {
		if ev.Key == "log" {
			count++
		}
	}
	if count > 0 {
		t.Fatal("cannot disable events")
	}
}

func TestUnitEmitFailureStartup(t *testing.T) {
	out := make(chan *eventRecord)
	emitter := newEventEmitter([]string{}, out)
	go func() {
		emitter.EmitFailureStartup("mocked error")
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "failure.startup" && ev.Value.Failure == "mocked error" {
			found = true
		}
	}
	if !found {
		t.Fatal("did not see expected event")
	}
}

func TestUnitEmitStatusProgress(t *testing.T) {
	out := make(chan *eventRecord)
	emitter := newEventEmitter([]string{}, out)
	go func() {
		emitter.EmitStatusProgress(0.7, "foo")
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "status.progress" && ev.Value.Message == "foo" && ev.Value.Percentage == 0.7 {
			found = true
		}
	}
	if !found {
		t.Fatal("did not see expected event")
	}
}

func TestUnitEmitNonblocking(t *testing.T) {
	out := make(chan *eventRecord)
	emitter := newEventEmitter([]string{}, out)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		emitter.Emit("log", eventValue{Message: "foo"})
		wg.Done()
	}()
	wg.Wait()
	if emitter.timeouts != 1 {
		t.Fatal("did not see any timeout")
	}
}
