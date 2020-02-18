// libooni exposes ooni/probe-engine as a C library. The API exposed
// by this library is API/ABI compatible with MK v0.10.x. Thus, it will
// allow MK integrators to easily migrate away from MK.
//
// See https://git.io/Jv4Rv (measurement-kit/measurement-kit @v0.10.9).
package main

// #include <limits.h>
// #include <stdint.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"

import (
	"encoding/json"
	"sync"

	"github.com/ooni/probe-engine/libooni/asynctask"
)

// TaskManager is a task manager
type TaskManager struct {
	index C.intptr_t
	lck   sync.Mutex
	tasks map[C.intptr_t]*asynctask.Task
}

// NewTaskManager creates a new TaskManager
func NewTaskManager() *TaskManager {
	return &TaskManager{index: 1, tasks: make(map[C.intptr_t]*asynctask.Task)}
}

// StartTask starts a new task. Returns a nonzero task
// handle on success, zero on failure.
func (tm *TaskManager) StartTask(csettings *C.char) C.intptr_t {
	if csettings == nil {
		return 0
	}
	gosettings := []byte(C.GoString(csettings))
	var settings asynctask.Settings
	if err := json.Unmarshal(gosettings, &settings); err != nil {
		return 0
	}
	taskptr, err := asynctask.StartTask(settings)
	if err != nil {
		return 0
	}
	tm.lck.Lock()
	handle := tm.index
	tm.index++
	tm.tasks[handle] = taskptr
	tm.lck.Unlock()
	return handle
}

// TaskWaitForNextEvent waits for the next event.
func (tm *TaskManager) TaskWaitForNextEvent(
	handle C.intptr_t, base **C.char, length *C.size_t,
) C.int {
	if handle == 0 || base == nil || length == nil {
		return 0
	}
	tm.lck.Lock()
	taskptr := tm.tasks[handle]
	tm.lck.Unlock()
	eventptr, err := taskptr.WaitForNextEvent() // gracefully handles nil
	if err != nil {
		return 0
	}
	if eventptr == nil {
		// Rationale: we used to return `null` when done. But then I figured that
		// this could break people code, because they need to write conditional
		// coding for handling both an ordinary event and `null`. So, to ease the
		// integrator's life, we now return a dummy, well formed event.
		eventptr = &asynctask.Event{Key: "task_terminated"}
	}
	data, err := json.Marshal(eventptr)
	if err != nil {
		return 0
	}
	sdata := string(data)
	if len(sdata) <= 0 || uint64(len(sdata)) > uint64(C.SIZE_MAX) {
		return 0
	}
	*base = C.CString(sdata)
	*length = C.size_t(len(sdata))
	return 1
}

// TaskIsDone returns whether the task is done
func (tm *TaskManager) TaskIsDone(handle C.intptr_t) C.int {
	tm.lck.Lock()
	taskptr := tm.tasks[handle]
	tm.lck.Unlock()
	if taskptr.IsDone() { // gracefully handles nil
		return 1
	}
	return 0
}

// TaskInterrupt interrupts a running task
func (tm *TaskManager) TaskInterrupt(handle C.intptr_t) {
	tm.lck.Lock()
	taskptr := tm.tasks[handle]
	tm.lck.Unlock()
	taskptr.Interrupt() // gracefully handles nil
}

// TaskDestroy destroys a task
func (tm *TaskManager) TaskDestroy(handle C.intptr_t) {
	tm.lck.Lock()
	taskptr := tm.tasks[handle]
	delete(tm.tasks, handle)
	tm.lck.Unlock()
	// the following code gracefully handles nil
	taskptr.Interrupt()
	for !taskptr.IsDone() {
		taskptr.WaitForNextEvent()
	}
}

var tm = NewTaskManager()

//export OONIGoTaskStart
func OONIGoTaskStart(csettings *C.char) C.intptr_t {
	return tm.StartTask(csettings)
}

//export OONIGoTaskWaitForNextEvent
func OONIGoTaskWaitForNextEvent(handle C.intptr_t, base **C.char, length *C.size_t) C.int {
	return tm.TaskWaitForNextEvent(handle, base, length)
}

//export OONIGoTaskIsDone
func OONIGoTaskIsDone(handle C.intptr_t) C.int {
	return tm.TaskIsDone(handle)
}

//export OONIGoTaskInterrupt
func OONIGoTaskInterrupt(handle C.intptr_t) {
	tm.TaskInterrupt(handle)
}

//export OONIGoTaskDestroy
func OONIGoTaskDestroy(handle C.intptr_t) {
	tm.TaskDestroy(handle)
}

func main() {}
