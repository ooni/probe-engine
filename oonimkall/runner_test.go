package oonimkall

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	engine "github.com/ooni/probe-engine"
)

func TestUnitRunnerHasUnsupportedSettings(t *testing.T) {
	out := make(chan *eventRecord)
	settings := &settingsRecord{
		InputFilepaths: []string{"foo"},
		Name:           "example",
		Options: settingsOptions{
			Backend:          "foo",
			CABundlePath:     "foo",
			GeoIPASNPath:     "foo",
			GeoIPCountryPath: "foo",
			NoFileReport:     false,
			ProbeASN:         "AS0",
			ProbeCC:          "ZZ",
			ProbeIP:          "127.0.0.1",
			ProbeNetworkName: "XXX",
			RandomizeInput:   true,
		},
		OutputFilepath: "foo",
	}
	go func() {
		defer close(out)
		r := newRunner(settings, out)
		logger := newChanLogger(r.emitter, "WARNING", r.out)
		if r.hasUnsupportedSettings(logger) != true {
			panic("expected to see unsupported settings")
		}
	}()
	var seen []string
	for ev := range out {
		switch ev.Key {
		case "failure.startup":
			if strings.HasSuffix("not supported", ev.Value.Failure) {
				log.Fatalf("invalid value: %s", ev.Value.Failure)
			}
			seen = append(seen, ev.Value.Failure)
		case "log":
			if strings.HasSuffix("not supported", ev.Value.Message) {
				log.Fatalf("invalid value: %s", ev.Value.Message)
			}
			seen = append(seen, ev.Value.Message)
		default:
			log.Fatalf("invalid key: %s", ev.Key)
		}
	}
	const expected = 11
	if len(seen) != expected {
		t.Fatalf("expected: %d; seen %+v", expected, seen)
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
