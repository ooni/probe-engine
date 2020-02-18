// Package asynctask implements measurement-kit's v0.10.9 API.
//
// See https://git.io/Jv4Rv (measurement-kit/measurement-kit@v0.10.9)
// for a comprehensive description of such MK API.
package asynctask

import (
	"encoding/json"
	"sync"

	"github.com/m-lab/go/rtx"
	"github.com/ooni/probe-engine/libooni/asynctask/internal"
)

// Task is an asynchronous task.
type Task struct {
	*internal.Task
}

// Start starts an asynchronous task. The input argument is a
// serialized JSON conforming to MK v0.10.9's API.
func Start(input string) (*Task, error) {
	var settings internal.Settings
	if err := json.Unmarshal([]byte(input), &settings); err != nil {
		return 0, err
	}
	taskptr, err := internal.StartTask(settings)
	if err != nil {
		return 0, err
	}
	return &Task{Task: taskptr}, nil
}

// WaitForNextEvent blocks until the next event occurs. The returned
// string is a serialized JSON following MK v0.10.9's API.
func (t *Task) WaitForNextEvent() string {
	const terminated = `{"key":"task_terminated","value":{}}` // like MK
	evp := t.Task.WaitForNextEvent()
	if evp == nil {
		return terminated
	}
	data, err := json.Marshal(evp)
	rtx.PanicOnError(err, "json.Marshal failed")
	return string(data)
}

// IsDone returns true if the task is done. 
func (t *Task) IsDone() bool {
	return t.Task.IsDone()
}

// Interrupt interrupts the task.
func Interrupt() {
	t.Task.Interrupt()
}
