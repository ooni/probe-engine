package oonimkall_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ooni/probe-engine/mobile/oonimkall"
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
		//t.Logf("%+v", event)
		fmt.Printf("%+v\n", event)
	}
	// continue draining until we see the status.terminated event
	for {
		eventstr := task.WaitForNextEvent()
		var event eventlike
		if err := json.Unmarshal([]byte(eventstr), &event); err != nil {
			t.Fatal(err)
		}
		if event.Key == "task_terminated" {
			break
		}
		fmt.Printf("%+v\n", event)
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
