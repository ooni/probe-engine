// Command apitool is a simple tool to fetch individual OONI measurements.
//
// This tool IS NOT intended for batch downloading.
//
// Please, see https://ooni.org/data for information pertaining how to
// access OONI data in bulk. Please see https://explorer.ooni.org if your
// intent is to navigate and explore OONI data
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/ooniapi"
)

func newclient() *ooniapi.Client {
	return &ooniapi.Client{
		BaseURL:    "https://api.ooni.io",
		HTTPClient: http.DefaultClient,
	}
}

var osExit = os.Exit

func fatalOnError(err error, message string) {
	if err != nil {
		log.WithError(err).Error(message)
		osExit(1) // overridable from tests
	}
}

var (
	debug    = flag.Bool("v", false, "Enable verbose mode")
	input    = flag.String("input", "", "Input of the measurement")
	mode     = flag.String("mode", "", "One of: check, meta, raw")
	reportid = flag.String("report-id", "", "Report ID of the measurement")
)

var logmap = map[bool]log.Level{
	true:  log.DebugLevel,
	false: log.InfoLevel,
}

func main() {
	flag.Parse()
	log.SetLevel(logmap[*debug])
	client := newclient()
	switch *mode {
	case "check":
		check(client)
	case "meta":
		meta(client)
	case "raw":
		raw(client)
	default:
		fatalOnError(fmt.Errorf("invalid -mode flag value: %s", *mode), "usage error")
	}
}

func check(c *ooniapi.Client) {
	resp, err := c.CheckReportID(context.Background(), &ooniapi.CheckReportIDRequest{
		ReportID: *reportid,
	})
	fatalOnError(err, "c.CheckReportID failed")
	fmt.Printf("%+v\n", resp.Found)
}

func meta(c *ooniapi.Client) {
	pprint(mmeta(c, false))
}

func raw(c *ooniapi.Client) {
	m := mmeta(c, true)
	rm := []byte(m.RawMeasurement)
	var opaque interface{}
	err := json.Unmarshal(rm, &opaque)
	fatalOnError(err, "json.Unmarshal failed")
	pprint(opaque)
}

func pprint(opaque interface{}) {
	data, err := json.MarshalIndent(opaque, "", "  ")
	fatalOnError(err, "json.MarshalIndent failed")
	fmt.Printf("%s\n", data)
}

func mmeta(c *ooniapi.Client, full bool) *ooniapi.MeasurementMetaResponse {
	req := &ooniapi.MeasurementMetaRequest{
		ReportID: *reportid,
		Full:     full,
		Input:    *input,
	}
	ctx := context.Background()
	m, err := c.MeasurementMeta(ctx, req)
	fatalOnError(err, "client.GetMeasurementMeta failed")
	return m
}
