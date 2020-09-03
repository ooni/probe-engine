package probeservices_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		TestKeys:             fakeTestKeys{Failure: nil},
		TestName:             rt.TestName,
		TestStartTime:        "2018-11-01 15:33:17",
		TestVersion:          rt.TestVersion,
	}
}

func TestNewReportTemplate(t *testing.T) {
	m := &model.Measurement{
		ProbeASN:        "AS117",
		ProbeCC:         "IT",
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.1.0",
		TestName:        "dummy",
		TestVersion:     "0.1.0",
	}
	rt := probeservices.NewReportTemplate(m)
	expect := probeservices.ReportTemplate{
		DataFormatVersion: probeservices.DefaultDataFormatVersion,
		Format:            probeservices.DefaultFormat,
		ProbeASN:          "AS117",
		ProbeCC:           "IT",
		SoftwareName:      "ooniprobe-engine",
		SoftwareVersion:   "0.1.0",
		TestName:          "dummy",
		TestVersion:       "0.1.0",
	}
	if diff := cmp.Diff(expect, rt); diff != "" {
		t.Fatal(diff)
	}
}

func TestReportLifecycle(t *testing.T) {
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
	client := newclient()
	report, err := client.OpenReport(ctx, template)
	if err != nil {
		t.Fatal(err)
	}
	measurement := makeMeasurement(template, report.ID)
	if report.CanSubmit(&measurement) != true {
		t.Fatal("report should be able to submit this measurement")
	}
	if err = report.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
	if err = report.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestReportLifecycleWrongExperiment(t *testing.T) {
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
	client := newclient()
	report, err := client.OpenReport(ctx, template)
	if err != nil {
		t.Fatal(err)
	}
	defer report.Close(ctx)
	measurement := makeMeasurement(template, report.ID)
	measurement.TestName = "antani"
	if report.CanSubmit(&measurement) != false {
		t.Fatal("report should not be able to submit this measurement")
	}
}

func TestOpenReportInvalidDataFormatVersion(t *testing.T) {
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
	client := newclient()
	report, err := client.OpenReport(ctx, template)
	if !errors.Is(err, probeservices.ErrUnsupportedDataFormatVersion) {
		t.Fatal("not the error we expected")
	}
	if report != nil {
		t.Fatal("expected a nil report here")
	}
}

func TestOpenReportInvalidFormat(t *testing.T) {
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
	client := newclient()
	report, err := client.OpenReport(ctx, template)
	if !errors.Is(err, probeservices.ErrUnsupportedFormat) {
		t.Fatal("not the error we expected")
	}
	if report != nil {
		t.Fatal("expected a nil report here")
	}
}

func TestJSONAPIClientCreateFailure(t *testing.T) {
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
	client := newclient()
	client.BaseURL = "\t" // breaks the URL parser
	report, err := client.OpenReport(ctx, template)
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
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
	client := newclient()
	client.BaseURL = server.URL
	report, err := client.OpenReport(ctx, template)
	if !errors.Is(err, probeservices.ErrJSONFormatNotSupported) {
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
	client := newclient()
	client.BaseURL = server.URL
	report, err := client.OpenReport(ctx, template)
	if err != nil {
		t.Fatal(err)
	}
	measurement := makeMeasurement(template, report.ID)
	if err = report.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
	if err = report.Close(ctx); err != nil {
		t.Fatal(err)
	}
}
