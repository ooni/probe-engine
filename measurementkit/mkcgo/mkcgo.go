// +build !nomk

// Package mkcgo contains CGO bindings to Measurement Kit.
package mkcgo

import (
	// #include <measurement_kit/common/version.h>
	// #include <measurement_kit/ffi.h>
	//
	// #include <stdlib.h>
	//
	// #cgo darwin,amd64 LDFLAGS: /usr/local/lib/libmeasurement_kit.a
	// #cgo darwin,amd64 LDFLAGS: /usr/local/opt/libevent/lib/libevent_core.a
	// #cgo darwin,amd64 LDFLAGS: /usr/local/opt/libevent/lib/libevent_extra.a
	// #cgo darwin,amd64 LDFLAGS: /usr/local/opt/libevent/lib/libevent_openssl.a
	// #cgo darwin,amd64 LDFLAGS: /usr/local/opt/libevent/lib/libevent_pthreads.a
	// #cgo darwin,amd64 LDFLAGS: /usr/local/opt/libmaxminddb/lib/libmaxminddb.a
	// #cgo darwin,amd64 LDFLAGS: /usr/local/opt/openssl@1.1/lib/libssl.a
	// #cgo darwin,amd64 LDFLAGS: /usr/local/opt/openssl@1.1/lib/libcrypto.a
	// #cgo darwin,amd64 LDFLAGS: -lcurl
	//
	// #cgo windows LDFLAGS: -static
	// #cgo windows,amd64 CFLAGS: -I/usr/local/opt/mingw-w64-measurement-kit/include/
	// #cgo windows,amd64 LDFLAGS: /usr/local/opt/mingw-w64-measurement-kit/lib/libmeasurement_kit.a
	// #cgo windows,amd64 LDFLAGS: /usr/local/opt/mingw-w64-libmaxminddb/lib/libmaxminddb.a
	// #cgo windows,amd64 LDFLAGS: /usr/local/opt/mingw-w64-libevent/lib/libevent_openssl.a
	// #cgo windows,amd64 LDFLAGS: /usr/local/opt/mingw-w64-libressl/lib/libssl.a
	// #cgo windows,amd64 LDFLAGS: /usr/local/opt/mingw-w64-libressl/lib/libcrypto.a
	// #cgo windows,amd64 LDFLAGS: /usr/local/opt/mingw-w64-libevent/lib/libevent_core.a
	// #cgo windows,amd64 LDFLAGS: /usr/local/opt/mingw-w64-libevent/lib/libevent_extra.a
	// #cgo windows,amd64 LDFLAGS: -lws2_32 -liphlpapi
	// #cgo windows,amd64 LDFLAGS: -fstack-protector-strong
	//
	// #cgo linux,amd64,ooni LDFLAGS: -static
	// #cgo linux,amd64,ooni LDFLAGS: /usr/local/lib/libmeasurement_kit.a
	// #cgo linux,amd64,ooni LDFLAGS: /usr/lib/libmaxminddb.a
	// #cgo linux,amd64,ooni LDFLAGS: /usr/lib/libevent_openssl.a
	// #cgo linux,amd64,ooni LDFLAGS: /usr/lib/libssl.a
	// #cgo linux,amd64,ooni LDFLAGS: /usr/lib/libcrypto.a
	// #cgo linux,amd64,ooni LDFLAGS: /usr/lib/libevent_core.a
	// #cgo linux,amd64,ooni LDFLAGS: /usr/lib/libevent_extra.a
	// #cgo linux,amd64,ooni LDFLAGS: /usr/lib/libevent_pthreads.a
	// #cgo linux,amd64,ooni LDFLAGS: /lib/libz.a
	//
	// #cgo linux,!ooni LDFLAGS: -lmeasurement_kit
	//
	// #if MK_VERSION_NUMERIC != 0x00000000010000121LL
	// #error "Wrong measurement-kit version, please recompile measurement-kit"
	// #endif
	"C"
	"errors"
	"unsafe"
)

func evprocess(taskp *C.mk_task_t, out chan<- []byte) {
	eventp := C.mk_task_wait_for_next_event(taskp)
	if eventp == nil {
		return
	}
	defer C.mk_event_destroy(eventp)
	events := C.mk_event_serialize(eventp)
	if events == nil {
		return
	}
	out <- []byte(C.GoString(events))
}

func taskloop(taskp *C.mk_task_t, out chan<- []byte) {
	defer close(out)
	defer C.mk_task_destroy(taskp)
	for C.mk_task_is_done(taskp) == 0 {
		evprocess(taskp, out)
	}
}

func taskstart(settings []byte) *C.mk_task_t {
	settingsp := C.CString(string(settings))
	if settingsp == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(settingsp))
	return C.mk_task_start(settingsp)
}

// Start starts a Measurement Kit task.
func Start(settings []byte) (<-chan []byte, error) {
	taskp := taskstart(settings)
	if taskp == nil {
		return nil, errors.New("C.mk_task_start failed")
	}
	out := make(chan []byte)
	go taskloop(taskp, out)
	return out, nil
}
