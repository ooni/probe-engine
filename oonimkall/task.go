// Package oonimkall implements measurement-kit's FFI API.
//
// See https://git.io/Jv4Rv (measurement-kit/measurement-kit@v0.10.9)
// for a comprehensive description of MK's FFI API.
//
// See also https://github.com/ooni/probe-engine/pull/347 for the
// design document describing this API.
//
// See also https://github.com/ooni/probe-engine/blob/master/DESIGN.md.
package oonimkall

import (
	"context"
	"encoding/json"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/runtimex"
)

// Task is an asynchronous task.
type Task struct {
	cancel    context.CancelFunc
	isdone    *atomicx.Int64
	isstopped *atomicx.Int64
	out       chan *eventRecord
}

// StartTask starts an asynchronous task. The input argument is a
// serialized JSON conforming to MK v0.10.9's API.
func StartTask(input string) (*Task, error) {
	var settings settingsRecord
	if err := json.Unmarshal([]byte(input), &settings); err != nil {
		return nil, err
	}
	const bufsiz = 128 // common case: we don't want runner to block
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		cancel:    cancel,
		isdone:    atomicx.NewInt64(),
		isstopped: atomicx.NewInt64(),
		out:       make(chan *eventRecord, bufsiz),
	}
	go func() {
		defer close(task.out)
		defer task.isstopped.Add(1)
		r := newRunner(&settings, task.out)
		r.Run(ctx)
	}()
	return task, nil
}

// WaitForNextEvent blocks until the next event occurs. The returned
// string is a serialized JSON following MK v0.10.9's API.
func (t *Task) WaitForNextEvent() string {
	const terminated = `{"key":"task_terminated","value":{}}` // like MK
	evp := <-t.out
	if evp == nil {
		t.isdone.Add(1)
		return terminated
	}
	data, err := json.Marshal(evp)
	runtimex.PanicOnError(err, "json.Marshal failed")
	return string(data)
}

// IsDone returns true if the task is done.
func (t *Task) IsDone() bool {
	return t.isdone.Load() != 0
}

// Interrupt interrupts the task.
func (t *Task) Interrupt() {
	t.cancel()
}
