// Package mkall implements the mkall API. This API is implemented
// for iOS at github.com/measurement-kit/mkall-ios. Android code instead
// is at github.com/measurement-kit/android-libs. Providing replacement
// code for such APIs is part of our plan to go beyond MK.
package mkall

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/m-lab/go/rtx"
	"github.com/ooni/probe-engine/measurementkit/mkall/asynctask"
)

// AsyncTask is a measurementkit async task.
type AsyncTask struct {
	inner *asynctask.Task
}

// StartAsyncTask stars a new AsyncTask.
func StartAsyncTask(serialized string) (*AsyncTask, error) {
	var settings asynctask.Settings
	if err := json.Unmarshal([]byte(serialized), &settings); err != nil {
		return nil, err
	}
	task, err := asynctask.StartTask(settings)
	if err != nil {
		return nil, err
	}
	return &AsyncTask{inner: task}, nil
}

// WaitForNextEvent waits for next event.
func (task *AsyncTask) WaitForNextEvent() (serialized string) {
	if task != nil {
		ev, err := task.inner.WaitForNextEvent()
		rtx.PanicOnError(err, "task.inner.WaitForNextEvent failed")
		data, err := json.Marshal(ev)
		rtx.PanicOnError(err, "json.Marshal failed")
		serialized = string(data)
	}
	return
}

// IsDone returns whether the task is done.
func (task *Task) IsDone() (done bool) {
	if task != nil {
		done = task.inner.IsDone()
	}
	return
}

// Interrupt interrupts the task.
func (task *Task) Interrupt() {
	if task != nil {
		task.inner.cancel()
	}
}

