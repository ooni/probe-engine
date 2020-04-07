package main

import (
	"testing"
)

func TestUnitTaskStartNullPointer(t *testing.T) {
	if miniooni_cgo_task_start(nil) != 0 {
		t.Fatal("expected nil result here")
	}
}

func TestUnitTaskStartInvalidJSON(t *testing.T) {
	settings := cstring("{")
	defer freestring(settings)
	if miniooni_cgo_task_start(settings) != 0 {
		t.Fatal("expected 0 result here")
	}
}

func TestUnitTaskWaitForNextEventInvalidHandle(t *testing.T) {
	if miniooni_cgo_task_wait_for_next_event(0) != nil {
		t.Fatal("expected nil result here")
	}
}

func TestUnitTaskIsDoneInvalidHandle(t *testing.T) {
	if miniooni_cgo_task_is_done(0) == 0 {
		t.Fatal("expected true-ish result here")
	}
}

func TestUnitTaskInterruptInvalidHandle(t *testing.T) {
	miniooni_cgo_task_interrupt(0) // mainly: we don't crash :^)
}

func TestUnitEventDestroyInvalidString(t *testing.T) {
	miniooni_cgo_event_destroy(nil) // mainly: we don't crash
}

func TestUnitTaskDestroyInvalidHandle(t *testing.T) {
	miniooni_cgo_task_destroy(0) // mainly: we don't crash
}

func TestIntegrationExampleNormalUsage(t *testing.T) {
	settings := cstring(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	defer freestring(settings)
	task := miniooni_cgo_task_start(settings)
	if task == 0 {
		t.Fatal("expected nonzero task here")
	}
	for miniooni_cgo_task_is_done(task) == 0 {
		event := miniooni_cgo_task_wait_for_next_event(task)
		t.Logf("%s", gostring(event))
		miniooni_cgo_event_destroy(event)
	}
	miniooni_cgo_task_destroy(task)
}

func TestIntegrationExampleInterruptAndDestroy(t *testing.T) {
	settings := cstring(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	defer freestring(settings)
	task := miniooni_cgo_task_start(settings)
	if task == 0 {
		t.Fatal("expected nonzero task here")
	}
	miniooni_cgo_task_interrupt(task)
	miniooni_cgo_task_destroy(task)
}

func TestIntegrationExampleDestroyImmediately(t *testing.T) {
	settings := cstring(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	defer freestring(settings)
	task := miniooni_cgo_task_start(settings)
	if task == 0 {
		t.Fatal("expected nonzero task here")
	}
	miniooni_cgo_task_destroy(task)
}

func TestUnitTaskStartIdxWrapping(t *testing.T) {
	settings := cstring(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	defer freestring(settings)
	o := setmaxidx()
	// do twice and see if it's idempotent
	if task := miniooni_cgo_task_start(settings); task != 0 {
		t.Fatal("expected zero task here")
	}
	if task := miniooni_cgo_task_start(settings); task != 0 {
		t.Fatal("expected zero task here")
	}
	restoreidx(o)
}
