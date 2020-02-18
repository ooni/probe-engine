// Package asynctask implements measurement-kit's async task API.
//
// See https://git.io/Jv4Rv (measurement-kit/measurement-kit@v0.10.9).
package asynctask

import (
	"encoding/json"
	"sync"

	"github.com/m-lab/go/rtx"
	"github.com/ooni/probe-engine/libooni/asynctask/internal"
)

var (
	idx int64 = 1
	m = make(map[int64]*internal.Task)
	mu sync.Mutex
)

// Start starts an OONI task.
func Start(input string) (int64, error) {
	var settings internal.Settings
	if err := json.Unmarshal([]byte(input), &settings); err != nil {
		return 0, err
	}
	taskptr, err := internal.StartTask(settings)
	if err != nil {
		return 0, err
	}
	mu.Lock()
	task := idx
	idx++
	m[task] = taskptr
	mu.Unlock()
	return task, nil
}

// WaitForNextEvent blocks until the next event occurs.
func WaitForNextEvent(task int64) string {
	mu.Lock()
	taskptr := m[task]
	mu.Unlock()
	const terminated = `{"key":"task_terminated","value":{}}`
	if taskptr == nil {
		return terminated
	}
	// TODO(bassosimone): no need to have error here
	evp, _ := taskptr.WaitForNextEvent()
	if evp == nil {
		return terminated
	}
	data, err := json.Marshal(evp)
	rtx.PanicOnError(err, "json.Marshal failed")
	return string(data)
}

// IsDone returns true if the task is done. 
func IsDone(task int64) (isdone bool) {
	mu.Lock()
	taskptr := m[task]
	isdone = (taskptr == nil || taskptr.IsDone())
	mu.Unlock()
	return
}

// Interrupt interrupts a running task.
func Interrupt(task int64) {
	mu.Lock()
	if taskptr := m[task]; taskptr != nil {
		taskptr.Interrupt()
	}
	mu.Unlock()
}

// Stop stops a running task.
func Stop(task int64) {
	mu.Lock()
	taskptr := m[task]
	mu.Unlock()
	if taskptr == nil {
		return
	}
	taskptr.Interrupt()
	for !taskptr.IsDone() {
		taskptr.WaitForNextEvent()
	}
}
