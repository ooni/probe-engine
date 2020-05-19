package probeservices_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/jsonapi"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/probeservices"
)

type fakeTestKeys struct {
	Failure *string `json:"failure"`
}

func makeMeasurement(rt probeservices.ReportTemplate, ID string) model.Measurement {
	return model.Measurement{
		DataFormatVersion:    probeservices.DefaultDataFormatVersion,
		ID:                   "bdd20d7a-bba5-40dd-a111-9863d7908572",
		MeasurementRuntime:   5.0565230846405,
		MeasurementStartTime: "2018-11-01 15:33:20",
		ProbeIP:              "1.2.3.4",
		ProbeASN:             rt.ProbeASN,
		ProbeCC:              rt.ProbeCC,
		ReportID:             ID,
		ResolverASN:          "AS15169",
		ResolverIP:           "8.8.8.8",
		ResolverNetworkName:  "Google LLC",
		SoftwareName:         rt.SoftwareName,
		SoftwareVersion:      rt.SoftwareVersion,
		TestKeys: fakeTestKeys{
			Failure: nil,
		},
		TestName:      rt.TestName,
		TestStartTime: "2018-11-01 15:33:17",
		TestVersion:   rt.TestVersion,
	}
}

func makeClient() probeservices.Client {
	return probeservices.Client{
		Client: jsonapi.Client{
			BaseURL:    "https://ps-test.ooni.io/",
			HTTPClient: http.DefaultClient,
			Logger:     log.Log,
			UserAgent:  "ooniprobe-engine/0.1.0",
		},
	}
}

func TestReportLifecycle(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := probeservices.ReportTemplate{
		DataFormatVersion: probeservices.DefaultDataFormatVersion,
		Format:            probeservices.DefaultFormat,
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
	template := probeservices.ReportTemplate{
		DataFormatVersion: "0.1.0",
		Format:            probeservices.DefaultFormat,
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
	template := probeservices.ReportTemplate{
		DataFormatVersion: probeservices.DefaultDataFormatVersion,
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
	template := probeservices.ReportTemplate{
		DataFormatVersion: probeservices.DefaultDataFormatVersion,
		Format:            probeservices.DefaultFormat,
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
	template := probeservices.ReportTemplate{
		DataFormatVersion: probeservices.DefaultDataFormatVersion,
		Format:            probeservices.DefaultFormat,
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
				sdata, err := ioutil.ReadFile("../testdata/collector-expected.jsonl")
				if err != nil {
					panic(err)
				}
				if !bytes.Equal(data, sdata) {
					panic("mismatch between submission and disk")
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
	template := probeservices.ReportTemplate{
		DataFormatVersion: probeservices.DefaultDataFormatVersion,
		Format:            probeservices.DefaultFormat,
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
