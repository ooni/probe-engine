package oonimkall

import (
	"fmt"
	"testing"
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
