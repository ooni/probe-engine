package oonimkall

import (
	"context"
	"errors"
	"fmt"
	"testing"

	engine "github.com/ooni/probe-engine"
)

func TestUnitRunnerHasUnsupportedSettings(t *testing.T) {
	var noFileReport, randomizeInput bool
	out := make(chan *eventRecord)
	settings := &settingsRecord{
		InputFilepaths: []string{"foo"},
		Options: settingsOptions{
			Backend:          "foo",
			CABundlePath:     "foo",
			GeoIPASNPath:     "foo",
			GeoIPCountryPath: "foo",
			NoFileReport:     &noFileReport,
			ProbeASN:         "AS0",
			ProbeCC:          "ZZ",
			ProbeIP:          "127.0.0.1",
			ProbeNetworkName: "XXX",
			RandomizeInput:   &randomizeInput,
		},
		OutputFilePath: "foo",
	}
	numseen := make(chan int)
	go func() {
		var count int
		for ev := range out {
			if ev.Key != "failure.startup" {
				panic(fmt.Sprintf("invalid key: %s", ev.Key))
			}
			count++
		}
		numseen <- count
	}()
	r := newRunner(settings, out)
	if r.hasUnsupportedSettings() != true {
		t.Fatal("expected to see unsupported settings")
	}
	close(out)
	const expected = 12
	if n := <-numseen; n != expected {
		t.Fatalf("expected: %d; seen %d", expected, n)
	}
}

func TestUnitMeasurementSubmissionEventName(t *testing.T) {
	if measurementSubmissionEventName(nil) != statusMeasurementSubmission {
		t.Fatal("unexpected submission event name")
	}
	if measurementSubmissionEventName(errors.New("mocked error")) != failureMeasurementSubmission {
		t.Fatal("unexpected submission event name")
	}
}

func TestUnitMeasurementSubmissionFailure(t *testing.T) {
	if measurementSubmissionFailure(nil) != "" {
		t.Fatal("unexpected submission failure")
	}
	if measurementSubmissionFailure(errors.New("mocked error")) != "mocked error" {
		t.Fatal("unexpected submission failure")
	}
}

func TestIntegrationRunnerMaybeLookupLocationFailure(t *testing.T) {
	out := make(chan *eventRecord)
	settings := &settingsRecord{
		AssetsDir: "../../testdata/oonimkall/assets",
		Name:      "Example",
		Options: settingsOptions{
			SoftwareName:    "oonimkall-test",
			SoftwareVersion: "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
		TempDir:  "../../testdata/oonimkall/tmp",
	}
	seench := make(chan int64)
	go func() {
		var seen int64
		for ev := range out {
			switch ev.Key {
			case "failure.ip_lookup", "failure.asn_lookup",
				"failure.cc_lookup", "failure.resolver_lookup":
				seen++
			case "status.progress":
				if ev.Value.Percentage >= 0.2 {
					panic(fmt.Sprintf("too much progress: %+v", ev))
				}
			case "status.queued", "status.started", "status.end":
			default:
				panic(fmt.Sprintf("unexpected key: %s", ev.Key))
			}
		}
		seench <- seen
	}()
	expected := errors.New("mocked error")
	r := newRunner(settings, out)
	r.maybeLookupLocation = func(*engine.Session) error {
		return expected
	}
	r.Run(context.Background())
	close(out)
	if n := <-seench; n != 4 {
		t.Fatal("unexpected number of events")
	}
}
