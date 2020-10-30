package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	engine "github.com/ooni/probe-engine"
)

func TestRunnerHasUnsupportedSettings(t *testing.T) {
	out := make(chan *Event)
	var falsebool bool
	var zerodotzero float64
	var zero int64
	var emptystring string
	settings := &Settings{
		InputFilepaths: []string{"foo"},
		Options: SettingsOptions{
			AllEndpoints:          &falsebool,
			Backend:               "foo",
			BouncerBaseURL:        "https://ps-nonexistent.ooni.io/",
			CABundlePath:          "foo",
			CollectorBaseURL:      "https://ps-nonexistent.ooni.io/",
			ConstantBitrate:       &falsebool,
			DNSNameserver:         &emptystring,
			DNSEngine:             &emptystring,
			ExpectedBody:          &emptystring,
			GeoIPASNPath:          "foo",
			GeoIPCountryPath:      "foo",
			Hostname:              &emptystring,
			IgnoreBouncerError:    &falsebool,
			IgnoreOpenReportError: &falsebool,
			MLabNSAddressFamily:   &emptystring,
			MLabNSBaseURL:         &emptystring,
			MLabNSCountry:         &emptystring,
			MLabNSMetro:           &emptystring,
			MLabNSPolicy:          &emptystring,
			MLabNSToolName:        &emptystring,
			NoFileReport:          false,
			Port:                  &zero,
			ProbeASN:              "AS0",
			ProbeCC:               "ZZ",
			ProbeIP:               "127.0.0.1",
			ProbeNetworkName:      "XXX",
			RandomizeInput:        true,
			SaveRealResolverIP:    &falsebool,
			Server:                &emptystring,
			TestSuite:             &zero,
			Timeout:               &zerodotzero,
			UUID:                  &emptystring,
		},
		OutputFilepath: "foo",
	}
	go func() {
		defer close(out)
		r := NewRunner(settings, out)
		logger := NewChanLogger(r.emitter, "WARNING", r.out)
		if r.hasUnsupportedSettings(logger) != true {
			panic("expected to see unsupported settings")
		}
	}()
	var fatal, warn []string
	for ev := range out {
		switch ev.Key {
		case "failure.startup":
			evv := ev.Value.(EventFailure)
			if strings.HasSuffix("not supported", evv.Failure) {
				log.Fatalf("invalid value: %s", evv.Failure)
			}
			fatal = append(fatal, evv.Failure)
		case "log":
			evv := ev.Value.(EventLog)
			if strings.HasSuffix("not supported", evv.Message) {
				log.Fatalf("invalid value: %s", evv.Message)
			}
			warn = append(warn, evv.Message)
		default:
			log.Fatalf("invalid key: %s", ev.Key)
		}
	}
	expectedFatal := []string{
		"InputFilepaths: not supported",
		"Options.Backend: not supported",
		"Options.BouncerBaseURL: not supported",
		"Options.CollectorBaseURL: not supported",
		"Options.Port: not supported",
		"Options.RandomizeInput: not supported",
		"Options.SaveRealResolverIP: not supported",
		"Options.Server: not supported",
		"Options.TestSuite: not supported",
		"Options.Timeout: not supported",
		"Options.UUID: not supported",
		"OutputFilepath && !NoFileReport: not supported",
	}
	if diff := cmp.Diff(expectedFatal, fatal); diff != "" {
		t.Fatal(diff)
	}
	expectedWarn := []string{
		"Options.AllEndpoints: not supported",
		"Options.CABundlePath: not supported",
		"Options.ConstantBitrate: not supported",
		"Options.DNSNameserver: not supported",
		"Options.DNSEngine: not supported",
		"Options.ExpectedBody: not supported",
		"Options.GeoIPASNPath: not supported",
		"Options.GeoIPCountryPath: not supported",
		"Options.Hostname: not supported",
		"Options.IgnoreBouncerError: not supported",
		"Options.IgnoreOpenReportError: not supported",
		"Options.MLabNSAddressFamily: not supported",
		"Options.MLabNSBaseURL: not supported",
		"Options.MLabNSCountry: not supported",
		"Options.MLabNSMetro: not supported",
		"Options.MLabNSPolicy: not supported",
		"Options.MLabNSToolName: not supported",
		"Options.ProbeASN: not supported",
		"Options.ProbeCC: not supported",
		"Options.ProbeIP: not supported",
		"Options.ProbeNetworkName: not supported",
	}
	if diff := cmp.Diff(expectedWarn, warn); diff != "" {
		t.Fatal(diff)
	}
}

func TestMeasurementSubmissionEventName(t *testing.T) {
	if measurementSubmissionEventName(nil) != statusMeasurementSubmission {
		t.Fatal("unexpected submission event name")
	}
	if measurementSubmissionEventName(errors.New("mocked error")) != failureMeasurementSubmission {
		t.Fatal("unexpected submission event name")
	}
}

func TestMeasurementSubmissionFailure(t *testing.T) {
	if measurementSubmissionFailure(nil) != "" {
		t.Fatal("unexpected submission failure")
	}
	if measurementSubmissionFailure(errors.New("mocked error")) != "mocked error" {
		t.Fatal("unexpected submission failure")
	}
}

func TestRunnerMaybeLookupLocationFailure(t *testing.T) {
	out := make(chan *Event)
	settings := &Settings{
		AssetsDir: "../../testdata/oonimkall/assets",
		Name:      "Example",
		Options: SettingsOptions{
			SoftwareName:    "oonimkall-test",
			SoftwareVersion: "0.1.0",
		},
		StateDir: "../../testdata/oonimkall/state",
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
				evv := ev.Value.(EventStatusProgress)
				if evv.Percentage >= 0.2 {
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
	r := NewRunner(settings, out)
	r.maybeLookupLocation = func(*engine.Session) error {
		return expected
	}
	r.Run(context.Background())
	close(out)
	if n := <-seench; n != 4 {
		t.Fatal("unexpected number of events")
	}
}
