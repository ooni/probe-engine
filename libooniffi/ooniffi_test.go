package main

import (
	"testing"
)

func TestUnitTaskStartNullPointer(t *testing.T) {
	if ooniffi_task_start_(nil) != 0 {
		t.Fatal("expected nil result here")
	}
}

func TestUnitTaskStartInvalidJSON(t *testing.T) {
	settings := cstring("{")
	defer ooniffi_string_free(settings)
	if ooniffi_task_start_(settings) != 0 {
		t.Fatal("expected 0 result here")
	}
}

func TestUnitTaskWaitForNextEventInvalidHandle(t *testing.T) {
	if ooniffi_task_yield_from(0) != nil {
		t.Fatal("expected nil result here")
	}
}

func TestUnitTaskIsDoneInvalidHandle(t *testing.T) {
	if ooniffi_task_done(0) == 0 {
		t.Fatal("expected true-ish result here")
	}
}

func TestUnitTaskInterruptInvalidHandle(t *testing.T) {
	ooniffi_task_interrupt(0) // mainly: we don't crash :^)
}

func TestUnitEventDestroyInvalidString(t *testing.T) {
	ooniffi_string_free(nil) // mainly: we don't crash
}

func TestUnitTaskDestroyInvalidHandle(t *testing.T) {
	ooniffi_task_destroy(0) // mainly: we don't crash
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
	defer ooniffi_string_free(settings)
	task := ooniffi_task_start_(settings)
	if task == 0 {
		t.Fatal("expected nonzero task here")
	}
	for ooniffi_task_done(task) == 0 {
		event := ooniffi_task_yield_from(task)
		t.Logf("%s", gostring(event))
		ooniffi_string_free(event)
	}
	ooniffi_task_destroy(task)
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
	defer ooniffi_string_free(settings)
	task := ooniffi_task_start_(settings)
	if task == 0 {
		t.Fatal("expected nonzero task here")
	}
	ooniffi_task_interrupt(task)
	ooniffi_task_destroy(task)
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
	defer ooniffi_string_free(settings)
	task := ooniffi_task_start_(settings)
	if task == 0 {
		t.Fatal("expected nonzero task here")
	}
	ooniffi_task_destroy(task)
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
	defer ooniffi_string_free(settings)
	o := setmaxidx()
	// do twice and see if it's idempotent
	if task := ooniffi_task_start_(settings); task != 0 {
		t.Fatal("expected zero task here")
	}
	if task := ooniffi_task_start_(settings); task != 0 {
		t.Fatal("expected zero task here")
	}
	restoreidx(o)
}
