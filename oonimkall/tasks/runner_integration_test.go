// +build integration

package tasks_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ooni/probe-engine/oonimkall/tasks"
)

func TestIntegrationRunnerMaybeLookupBackendsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		Name:      "Example",
		Options: tasks.SettingsOptions{
			ProbeServicesBaseURL: server.URL,
			SoftwareName:         "oonimkall-test",
			SoftwareVersion:      "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	seench := make(chan int64)
	go func() {
		var seen int64
		for ev := range out {
			switch ev.Key {
			case "failure.startup":
				seen++
			case "status.queued", "status.started", "log", "status.end":
			default:
				panic(fmt.Sprintf("unexpected key: %s", ev.Key))
			}
		}
		seench <- seen
	}()
	tasks.Run(context.Background(), settings, out)
	close(out)
	if n := <-seench; n != 1 {
		t.Fatal("unexpected number of events")
	}
}

func TestIntegrationRunnerOpenReportFailure(t *testing.T) {
	var (
		nreq int64
		mu   sync.Mutex
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		nreq++
		if nreq == 1 {
			w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(500)
	}))
	defer server.Close()
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		Name:      "Example",
		Options: tasks.SettingsOptions{
			ProbeServicesBaseURL: server.URL,
			SoftwareName:         "oonimkall-test",
			SoftwareVersion:      "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	seench := make(chan int64)
	go func() {
		var seen int64
		for ev := range out {
			switch ev.Key {
			case "failure.report_create":
				seen++
			case "status.progress":
				evv := ev.Value.(tasks.EventStatusProgress)
				if evv.Percentage >= 0.4 {
					panic(fmt.Sprintf("too much progress: %+v", ev))
				}
			case "status.queued", "status.started", "log", "status.end",
				"status.geoip_lookup", "status.resolver_lookup":
			default:
				panic(fmt.Sprintf("unexpected key: %s", ev.Key))
			}
		}
		seench <- seen
	}()
	tasks.Run(context.Background(), settings, out)
	close(out)
	if n := <-seench; n != 1 {
		t.Fatal("unexpected number of events")
	}
}

func TestIntegrationRunnerGood(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		LogLevel:  "DEBUG",
		Name:      "Example",
		Options: tasks.SettingsOptions{
			SoftwareName:    "oonimkall-test",
			SoftwareVersion: "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "status.end" {
			found = true
		}
	}
	if !found {
		t.Fatal("status.end event not found")
	}
}

func TestIntegrationRunnerWithUnsupportedSettings(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		LogLevel:  "DEBUG",
		Name:      "Example",
		Options: tasks.SettingsOptions{
			SoftwareName:    "oonimkall-test",
			SoftwareVersion: "0.1.0",
		},
		OutputFilepath: "/nonexistent",
		StateDir:       "../../testdata/oonimkall/state",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "failure.startup" {
			found = true
		}
	}
	if !found {
		t.Fatal("failure.startup event not found")
	}
}

func TestIntegrationRunnerWithInvalidKVStorePath(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		LogLevel:  "DEBUG",
		Name:      "Example",
		Options: tasks.SettingsOptions{
			SoftwareName:    "oonimkall-test",
			SoftwareVersion: "0.1.0",
		},
		StateDir: "/nonexistent/long/directory/name",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "failure.startup" {
			found = true
		}
	}
	if !found {
		t.Fatal("failure.startup event not found")
	}
}

func TestIntegrationRunnerWithInvalidExperimentName(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		LogLevel:  "DEBUG",
		Name:      "Nonexistent",
		Options: tasks.SettingsOptions{
			SoftwareName:    "oonimkall-test",
			SoftwareVersion: "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "failure.startup" {
			found = true
		}
	}
	if !found {
		t.Fatal("failure.startup event not found")
	}
}

func TestIntegrationRunnerWithInconsistentGeolookupSettings(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		LogLevel:  "DEBUG",
		Name:      "Example",
		Options: tasks.SettingsOptions{
			NoGeoIP:          true,
			NoResolverLookup: false,
			SoftwareName:     "oonimkall-test",
			SoftwareVersion:  "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "failure.startup" {
			found = true
		}
	}
	if !found {
		t.Fatal("failure.startup event not found")
	}
}

func TestIntegrationRunnerWithNoGeolookup(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		LogLevel:  "DEBUG",
		Name:      "Example",
		Options: tasks.SettingsOptions{
			NoGeoIP:          true,
			NoResolverLookup: true,
			SoftwareName:     "oonimkall-test",
			SoftwareVersion:  "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "status.end" {
			found = true
		}
	}
	if !found {
		t.Fatal("status.end event not found")
	}
}

func TestIntegrationRunnerWithMissingInput(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		LogLevel:  "DEBUG",
		Name:      "ExampleWithInput",
		Options: tasks.SettingsOptions{
			NoGeoIP:          true,
			NoResolverLookup: true,
			SoftwareName:     "oonimkall-test",
			SoftwareVersion:  "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "failure.startup" {
			found = true
		}
	}
	if !found {
		t.Fatal("failure.startup event not found")
	}
}

func TestIntegrationRunnerWithMaxRuntime(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		Inputs:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
		LogLevel:  "DEBUG",
		Name:      "ExampleWithInput",
		Options: tasks.SettingsOptions{
			MaxRuntime:       1,
			NoGeoIP:          true,
			NoResolverLookup: true,
			SoftwareName:     "oonimkall-test",
			SoftwareVersion:  "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	begin := time.Now()
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "status.end" {
			found = true
		}
	}
	if !found {
		t.Fatal("status.end event not found")
	}
	// The runtime is long because of ancillary operations and is even more
	// longer because of self shaping we may be performing (especially in
	// CI builds) using `-tags shaping`). We have experimentally determined
	// that ~10 seconds is the typical CI test run time. See:
	//
	// 1. https://github.com/ooni/probe-engine/pull/588/checks?check_run_id=667263788
	//
	// 2. https://github.com/ooni/probe-engine/pull/588/checks?check_run_id=667263855
	if time.Now().Sub(begin) > 10*time.Second {
		t.Fatal("expected shorter runtime")
	}
}

func TestIntegrationRunnerWithMaxRuntimeNonInterruptible(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		Inputs:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
		LogLevel:  "DEBUG",
		Name:      "ExampleWithInputNonInterruptible",
		Options: tasks.SettingsOptions{
			MaxRuntime:       1,
			NoGeoIP:          true,
			NoResolverLookup: true,
			SoftwareName:     "oonimkall-test",
			SoftwareVersion:  "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	begin := time.Now()
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "status.end" {
			found = true
		}
	}
	if !found {
		t.Fatal("status.end event not found")
	}
	// The runtime is long because of ancillary operations and is even more
	// longer because of self shaping we may be performing (especially in
	// CI builds) using `-tags shaping`). We have experimentally determined
	// that ~10 seconds is the typical CI test run time. See:
	//
	// 1. https://github.com/ooni/probe-engine/pull/588/checks?check_run_id=667263788
	//
	// 2. https://github.com/ooni/probe-engine/pull/588/checks?check_run_id=667263855
	if time.Now().Sub(begin) > 10*time.Second {
		t.Fatal("expected shorter runtime")
	}
}

func TestIntegrationRunnerWithFailedMeasurement(t *testing.T) {
	out := make(chan *tasks.Event)
	settings := &tasks.Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		Inputs:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
		LogLevel:  "DEBUG",
		Name:      "ExampleWithFailure",
		Options: tasks.SettingsOptions{
			MaxRuntime:       1,
			NoGeoIP:          true,
			NoResolverLookup: true,
			SoftwareName:     "oonimkall-test",
			SoftwareVersion:  "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
	}
	go func() {
		tasks.Run(context.Background(), settings, out)
		close(out)
	}()
	var found bool
	for ev := range out {
		if ev.Key == "failure.measurement" {
			found = true
		}
	}
	if !found {
		t.Fatal("failure.measurement event not found")
	}
}
