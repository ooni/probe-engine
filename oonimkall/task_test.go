package oonimkall_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ooni/probe-engine/oonimkall"
)

type eventlike struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func TestIntegrationGood(t *testing.T) {
	task, err := oonimkall.StartTask(`{
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
	if err != nil {
		t.Fatal(err)
	}
	// interrupt the task so we also exercise this functionality
	go func() {
		<-time.After(time.Second)
		task.Interrupt()
	}()
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", event)
	}
	// make sure we only see task_terminated at this point
	for {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key != "task_terminated" {
			t.Fatalf("unexpected event.Key: %s", event.Key)
		}
		break
	}
}

func TestIntegrationGoodWithoutGeoIPLookup(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"no_geoip": true,
			"no_resolver_lookup": true,
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", event)
	}
}

func TestIntegrationWithMeasurementFailure(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "ExampleWithFailure",
		"options": {
			"no_geoip": true,
			"no_resolver_lookup": true,
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", event)
	}
}

func TestIntegrationInvalidJSON(t *testing.T) {
	task, err := oonimkall.StartTask(`{`)
	var syntaxerr *json.SyntaxError
	if !errors.As(err, &syntaxerr) {
		t.Fatal("not the expected error")
	}
	if task != nil {
		t.Fatal("task is not nil")
	}
}

func TestIntegrationUnsupportedSetting(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"input_filepaths": ["/nonexistent"],
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.startup" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.startup")
	}
}

func TestIntegrationEmptyStateDir(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.startup" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.startup")
	}
}

func TestIntegrationEmptyAssetsDir(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.startup" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.startup")
	}
}

func TestIntegrationUnknownExperiment(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Antani",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.startup" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.startup")
	}
}

func TestIntegrationInvalidBouncerBaseURL(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"bouncer_base_url": "\t",
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.startup" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.startup")
	}
}

func TestIntegrationInconsistentGeoIPSettings(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"no_geoip": true,
			"no_resolver_lookup": false,
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.startup" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.startup")
	}
}

func TestIntegrationInputIsRequired(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "ExampleWithInput",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.startup" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.startup")
	}
}

func TestIntegrationBadCollectorURL(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"log_level": "DEBUG",
		"name": "Example",
		"options": {
			"collector_base_url": "\t",
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var seen bool
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "failure.report_create" {
			seen = true
		}
		t.Logf("%+v", event)
	}
	if !seen {
		t.Fatal("did not see failure.report_create")
	}
}

func TestIntegrationInterruptWebConnectivity(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"inputs": [
			"http://www.kernel.org/",
			"http://www.x.org/",
			"http://www.microsoft.com/",
			"http://www.slashdot.org/",
			"http://www.repubblica.it/",
			"http://www.google.it/",
			"http://ooni.org/"
		],
		"name": "WebConnectivity",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		switch event.Key {
		case "status.measurement_start":
			go task.Interrupt()
		}
		keys = append(keys, event.Key)
	}
	expect := []string{
		"status.queued",
		"status.started",
		"status.progress",
		"status.progress",
		"status.progress",
		"status.geoip_lookup",
		"status.resolver_lookup",
		"status.progress",
		"status.report_create",
		"status.end",
		"task_terminated",
	}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatal("seen different keys than expected")
	}
}

func TestIntegrationInterruptNdt7(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"name": "Ndt7",
		"options": {
			"software_name": "oonimkall-test",
			"software_version": "0.1.0"
		},
		"state_dir": "../../testdata/oonimkall/state",
		"temp_dir": "../../testdata/oonimkall/tmp"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		<-time.After(7 * time.Second)
		task.Interrupt()
	}()
	var keys []string
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, event.Key)
	}
	t.Log(keys)
	expect := []string{
		"status.queued",
		"status.started",
		"status.progress",
		"status.progress",
		"status.progress",
		"status.geoip_lookup",
		"status.resolver_lookup",
		"status.progress",
		"status.report_create",
		"status.end",
		"task_terminated",
	}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatal("seen different keys than expected")
	}
}
