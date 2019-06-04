package collector_test

import (
	"context"
	"net/http"
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
		DataFormatVersion:    "0.2.0",
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
		BaseURL:    "https://a.collector.ooni.io/",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}
}

func TestReportLifecycle(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	template := collector.ReportTemplate{
		ProbeASN:        "AS0",
		ProbeCC:         "ZZ",
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.1.0",
		TestName:        "dummy",
		TestVersion:     "0.1.0",
	}
	client := makeClient()
	report, err := client.OpenReport(ctx, template)
	if err != nil {
		t.Fatal(err)
	}
	defer report.Close(ctx)
	measurement := makeMeasurement(template, report.ID)
	err = report.SubmitMeasurement(ctx, &measurement)
	if err != nil {
		t.Fatal(err)
	}
}
