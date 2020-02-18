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
	"sync"
	"sync/atomic"

	"github.com/m-lab/go/rtx"
)

// Task is an asynchronous task.
type Task struct {
	cancel context.CancelFunc
	out    chan *eventRecord
	isdone int64
	wg     *sync.WaitGroup
}

// StartTask starts an asynchronous task. The input argument is a
// serialized JSON conforming to MK v0.10.9's API.
func StartTask(input string) (*Task, error) {
	var settings settingsRecord
	if err := json.Unmarshal([]byte(input), &settings); err != nil {
		return nil, err
	}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		cancel: cancel,
		out:    make(chan *eventRecord),
		wg:     wg,
	}
	go func() {
		defer close(task.out)
		defer wg.Done()
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
		atomic.AddInt64(&t.isdone, 1)
		return terminated
	}
	data, err := json.Marshal(evp)
	rtx.PanicOnError(err, "json.Marshal failed")
	return string(data)
}

// IsDone returns true if the task is done.
func (t *Task) IsDone() bool {
	return atomic.LoadInt64(&t.isdone) != 0
}

// Interrupt interrupts the task.
func (t *Task) Interrupt() {
	t.cancel()
}
