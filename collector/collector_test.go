package collector_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/collector"
	"github.com/ooni/probe-engine/model"
)

type fakeTestKeys struct {
	ClientResolver string `json:"client_resolver"`
}

func makeMeasurement(rt collector.ReportTemplate, ID string) model.Measurement {
	return model.Measurement{
		DataFormatVersion:    "0.3.0",
		ID:                   "bdd20d7a-bba5-40dd-a111-9863d7908572",
		MeasurementStartTime: "2018-11-01 15:33:20",
		ProbeIP:              "1.2.3.4",
		ProbeASN:             rt.ProbeASN,
		ProbeCC:              rt.ProbeCC,
		ReportID:             ID,
		SoftwareName:         rt.SoftwareName,
		SoftwareVersion:      rt.SoftwareVersion,
		TestKeys: fakeTestKeys{
			ClientResolver: "91.80.37.104",
		},
		TestName:           rt.TestName,
		MeasurementRuntime: 5.0565230846405,
		TestStartTime:      "2018-11-01 15:33:17",
		TestVersion:        rt.TestVersion,
	}
}

func makeClient() *collector.Client {
	return &collector.Client{
		BaseURL:    "https://ps-test.ooni.io/",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}
}

func TestReportLifecycle(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := collector.ReportTemplate{
		DataFormatVersion: collector.DefaultDataFormatVersion,
		Format:            collector.DefaultFormat,
		ProbeASN:          "AS0",
		ProbeCC:           "ZZ",
		SoftwareName:      "ooniprobe-engine",
		SoftwareVersion:   "0.1.0",
		TestName:          "dummy",
		TestVersion:       "0.1.0",
	}
	client := makeClient()
	report, err := client.OpenReport(ctx, template)
	if err != nil {
		t.Fatal(err)
	}
	measurement := makeMeasurement(template, report.ID)
	err = report.SubmitMeasurement(ctx, &measurement)
	if err != nil {
		t.Fatal(err)
	}
	err = report.Close(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOpenReportInvalidDataFormatVersion(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := collector.ReportTemplate{
		DataFormatVersion: "0.1.0",
		Format:            collector.DefaultFormat,
		ProbeASN:          "AS0",
		ProbeCC:           "ZZ",
		SoftwareName:      "ooniprobe-engine",
		SoftwareVersion:   "0.1.0",
		TestName:          "dummy",
		TestVersion:       "0.1.0",
	}
	client := makeClient()
	report, err := client.OpenReport(ctx, template)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if report != nil {
		t.Fatal("expected a nil report here")
	}
}

func TestOpenReportInvalidFormat(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := collector.ReportTemplate{
		DataFormatVersion: collector.DefaultDataFormatVersion,
		Format:            "yaml",
		ProbeASN:          "AS0",
		ProbeCC:           "ZZ",
		SoftwareName:      "ooniprobe-engine",
		SoftwareVersion:   "0.1.0",
		TestName:          "dummy",
		TestVersion:       "0.1.0",
	}
	client := makeClient()
	report, err := client.OpenReport(ctx, template)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if report != nil {
		t.Fatal("expected a nil report here")
	}
}

func TestJSONAPIClientCreateFailure(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := collector.ReportTemplate{
		DataFormatVersion: collector.DefaultDataFormatVersion,
		Format:            collector.DefaultFormat,
		ProbeASN:          "AS0",
		ProbeCC:           "ZZ",
		SoftwareName:      "ooniprobe-engine",
		SoftwareVersion:   "0.1.0",
		TestName:          "dummy",
		TestVersion:       "0.1.0",
	}
	client := makeClient()
	client.BaseURL = "\t" // breaks the URL parser
	report, err := client.OpenReport(ctx, template)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if report != nil {
		t.Fatal("expected a nil report here")
	}
}

func TestOpenResponseNoJSONSupport(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			writer.Write([]byte(`{"ID":"abc","supported_formats":["yaml"]}`))
		}),
	)
	defer server.Close()
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := collector.ReportTemplate{
		DataFormatVersion: collector.DefaultDataFormatVersion,
		Format:            collector.DefaultFormat,
		ProbeASN:          "AS0",
		ProbeCC:           "ZZ",
		SoftwareName:      "ooniprobe-engine",
		SoftwareVersion:   "0.1.0",
		TestName:          "dummy",
		TestVersion:       "0.1.0",
	}
	client := makeClient()
	client.BaseURL = server.URL
	report, err := client.OpenReport(ctx, template)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if report != nil {
		t.Fatal("expected a nil report here")
	}
}

func TestEndToEnd(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/report" {
				w.Write([]byte(`{"report_id":"_id","supported_formats":["json"]}`))
				return
			}
			if r.RequestURI == "/report/_id" {
				data, err := ioutil.ReadAll(r.Body)
				if err != nil {
					panic(err)
				}
				sdata := string(data)
				// TODO(bassosimone): could this be a fixture?
				const expectedString = `{"format":"json","content":{"data_format_version":"0.3.0","id":"bdd20d7a-bba5-40dd-a111-9863d7908572","measurement_start_time":"2018-11-01 15:33:20","test_runtime":5.0565230846405,"probe_asn":"AS0","probe_cc":"ZZ","probe_ip":"1.2.3.4","report_id":"_id","resolver_ip":"","software_name":"ooniprobe-engine","software_version":"0.1.0","test_keys":{"client_resolver":"91.80.37.104"},"test_name":"dummy","test_start_time":"2018-11-01 15:33:17","test_version":"0.1.0"}}`
				if sdata != expectedString {
					panic(sdata)
				}
				w.Write([]byte(`{"measurement_id":"e00c584e6e9e5326"}`))
				return
			}
			if r.RequestURI == "/report/_id/close" {
				w.Write([]byte(`{}`))
				return
			}
			panic(r.RequestURI)
		}),
	)
	defer server.Close()
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := collector.ReportTemplate{
		DataFormatVersion: collector.DefaultDataFormatVersion,
		Format:            collector.DefaultFormat,
		ProbeASN:          "AS0",
		ProbeCC:           "ZZ",
		SoftwareName:      "ooniprobe-engine",
		SoftwareVersion:   "0.1.0",
		TestName:          "dummy",
		TestVersion:       "0.1.0",
	}
	client := makeClient()
	client.BaseURL = server.URL
	report, err := client.OpenReport(ctx, template)
	if err != nil {
		t.Fatal(err)
	}
	measurement := makeMeasurement(template, report.ID)
	err = report.SubmitMeasurement(ctx, &measurement)
	if err != nil {
		t.Fatal(err)
	}
	err = report.Close(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
