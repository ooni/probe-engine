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
	Key   string                 `json:"key"`
	Value map[string]interface{} `json:"value"`
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

func TestIntegrationMaxRuntime(t *testing.T) {
	begin := time.Now()
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
		"inputs": ["a", "b", "c"],
		"name": "ExampleWithInput",
		"options": {
			"max_runtime": 1,
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
		task.WaitForNextEvent()
	}
	if time.Now().Sub(begin) > 4*time.Second {
		t.Fatal("expected shorter runtime")
	}
}

func TestIntegrationInterruptExampleWithInput(t *testing.T) {
	// We cannot use WebConnectivity until it's written in Go since
	// measurement-kit may not always be available
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
		"name": "ExampleWithInputNonInterruptible",
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
		// We compress the keys. What matters is basically that we
		// see just one of the many possible measurements here.
		if keys == nil || keys[len(keys)-1] != event.Key {
			keys = append(keys, event.Key)
		}
	}
	expect := []string{
		"status.queued",
		"status.started",
		"status.progress",
		"status.geoip_lookup",
		"status.resolver_lookup",
		"status.progress",
		"status.report_create",
		"status.measurement_start",
		"log",
		"status.progress",
		"measurement",
		"status.measurement_submission",
		"status.measurement_done",
		"status.end",
		"task_terminated",
	}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatalf("seen different keys than expected: %+v", keys)
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
		// We compress the keys because we don't know how many
		// status.progress we will see. What matters is that we
		// don't see a measurement submission, since it means
		// that we have interrupted the measurement.
		if keys == nil || keys[len(keys)-1] != event.Key {
			keys = append(keys, event.Key)
		}
	}
	expect := []string{
		"status.queued",
		"status.started",
		"status.progress",
		"status.geoip_lookup",
		"status.resolver_lookup",
		"status.progress",
		"status.report_create",
		"status.measurement_start",
		"status.progress",
		"status.end",
		"task_terminated",
	}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatal("seen different keys than expected")
	}
}

func TestIntegrationCountBytesForExample(t *testing.T) {
	task, err := oonimkall.StartTask(`{
		"assets_dir": "../../testdata/oonimkall/assets",
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
	var downloadKB, uploadKB float64
	for !task.IsDone() {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		switch event.Key {
		case "status.end":
			t.Log(eventstr)
			downloadKB = event.Value["downloaded_kb"].(float64)
			uploadKB = event.Value["uploaded_kb"].(float64)
		}
	}
	if downloadKB == 0 {
		t.Fatal("downloadKB is zero")
	}
	if uploadKB == 0 {
		t.Fatal("uploadKB is zero")
	}
}
