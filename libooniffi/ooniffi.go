package main

import (
	//#cgo CFLAGS: -I${SRCDIR}/../..
	//#include <stdint.h>
	//#include <stdlib.h>
	"C"
	"sync"
	"unsafe"

	"github.com/ooni/probe-engine/oonimkall"
)

var (
	idx C.intptr_t = 1
	m              = make(map[C.intptr_t]*oonimkall.Task)
	mu  sync.Mutex
)

func cstring(s string) *C.char {
	return C.CString(s)
}

//export ooniffi_string_free
func ooniffi_string_free(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func gostring(s *C.char) string {
	return C.GoString(s)
}

const maxIdx = C.INTPTR_MAX - 1

//export ooniffi_task_start_
func ooniffi_task_start_(settings *C.char) C.intptr_t {
	if settings == nil {
		return 0
	}
	tp, err := oonimkall.StartTask(gostring(settings))
	if err != nil {
		return 0
	}
	mu.Lock()
	defer mu.Unlock()
	// TODO(bassosimone): the following if is basic protection against
	// undefined behaviour, i.e., the counter wrapping around. A much
	// better strategy would probably be to restart from 1. However it's
	// also unclear if any device could run that many tests, so...
	if idx >= maxIdx {
		return 0
	}
	handle := idx
	idx++
	m[handle] = tp
	return handle
}

func setmaxidx() C.intptr_t {
	o := idx
	idx = maxIdx
	return o
}

func restoreidx(v C.intptr_t) {
	idx = v
}

//export ooniffi_task_yield_from
func ooniffi_task_yield_from(handle C.intptr_t) (event *C.char) {
	mu.Lock()
	tp := m[handle]
	mu.Unlock()
	if tp != nil {
		event = cstring(tp.WaitForNextEvent())
	}
	return
}

//export ooniffi_task_done
func ooniffi_task_done(handle C.intptr_t) C.int {
	var isdone C.int = 1
	mu.Lock()
	if tp := m[handle]; tp != nil && !tp.IsDone() {
		isdone = 0
	}
	mu.Unlock()
	return isdone
}

//export ooniffi_task_interrupt
func ooniffi_task_interrupt(handle C.intptr_t) {
	mu.Lock()
	if tp := m[handle]; tp != nil {
		tp.Interrupt()
	}
	mu.Unlock()
}

//export ooniffi_task_destroy
func ooniffi_task_destroy(handle C.intptr_t) {
	mu.Lock()
	tp := m[handle]
	delete(m, handle)
	mu.Unlock()
	if tp != nil { // drain task if needed
		tp.Interrupt()
		go func() {
			for !tp.IsDone() {
				tp.WaitForNextEvent()
			}
		}()
	}
}

func main() {}
