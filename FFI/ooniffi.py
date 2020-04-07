#!/usr/bin/env python3

import ctypes
import json
import os
import sys
import tempfile

def main():
    if len(sys.argv) != 3:
        sys.exit("usage: ./FFI/ooniffi.py <DLL> <settings-file>")
    with open(sys.argv[2], "rb") as filep:
        settings = json.load(filep)
    print(settings)
    with tempfile.TemporaryDirectory() as tmpdir:
        settings["assets_dir"] = os.sep.join([tmpdir, "assets"])
        settings["state_dir"] = os.sep.join([tmpdir, "state"])
        settings["temp_dir"] = os.sep.join([tmpdir, "tmp"])
    print(settings)
    settings = json.dumps(settings).encode("utf-8")
    dll = ctypes.CDLL(sys.argv[1])
    print(dll)
    dll.ooniffi_task_start.argtypes = [ctypes.c_char_p]
    dll.ooniffi_task_start.restype = ctypes.c_void_p
    task = dll.ooniffi_task_start(settings)
    print(task)
    dll.ooniffi_task_is_done.argtypes = [ctypes.c_void_p]
    dll.ooniffi_task_is_done.restype = ctypes.c_int
    dll.ooniffi_task_wait_for_next_event.argtypes = [ctypes.c_void_p]
    dll.ooniffi_task_wait_for_next_event.restype = ctypes.c_void_p
    dll.ooniffi_event_serialization.argtypes = [ctypes.c_void_p]
    dll.ooniffi_event_serialization.restype = ctypes.c_char_p
    dll.ooniffi_event_destroy.argtypes = [ctypes.c_void_p]
    while not dll.ooniffi_task_is_done(task):
        ev = dll.ooniffi_task_wait_for_next_event(task)
        print(dll.ooniffi_event_serialization(ev))
        dll.ooniffi_event_destroy(ev)
    dll.ooniffi_task_destroy.argtypes = [ctypes.c_void_p]
    dll.ooniffi_task_destroy(task)

if __name__ == "__main__":
    main()
