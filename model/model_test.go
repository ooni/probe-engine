package model_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ooni/probe-engine/model"
)

type fakeTestKeys struct {
	ClientResolver string `json:"client_resolver"`
	Body           string `json:"body"`
}

func makeMeasurement(probeIP, probeASN, probeCC string) model.Measurement {
	return model.Measurement{
		DataFormatVersion:    "0.2.0",
		ID:                   "bdd20d7a-bba5-40dd-a111-9863d7908572",
		MeasurementStartTime: "2018-11-01 15:33:20",
		ProbeIP:              probeIP,
		ProbeASN:             probeASN,
		ProbeCC:              probeCC,
		ReportID:             "",
		SoftwareName:         "probe-engine",
		SoftwareVersion:      "0.1.0",
		TestKeys: fakeTestKeys{
			ClientResolver: "91.80.37.104",
			Body: fmt.Sprintf(`
				<HTML><HEAD><TITLE>Your IP is %s</TITLE></HEAD>
				<BODY><P>Hey you, I see your IP and it's %s!</P></BODY>
			`, probeIP, probeIP),
		},
		TestName:           "dummy",
		MeasurementRuntime: 5.0565230846405,
		TestStartTime:      "2018-11-01 15:33:17",
		TestVersion:        "0.1.0",
	}
}

func TestScrubCommonCase(t *testing.T) {
	const probeIP = "130.192.91.211"
	const probeASN = "AS137"
	const probeCC = "IT"
	m := makeMeasurement(probeIP, probeASN, probeCC)
	privacy := model.PrivacySettings{
		IncludeCountry: true,
		IncludeASN:     true,
	}
	err := privacy.Apply(&m, model.LocationInfo{
		ProbeIP: probeIP, // minimal initialization required by Apply
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.ProbeASN != probeASN {
		t.Fatal("ProbeASN has been scrubbed")
	}
	if m.ProbeCC != probeCC {
		t.Fatal("ProbeCC has been scrubbed")
	}
	if m.ProbeIP == probeIP {
		t.Fatal("ProbeIP has not been scrubbed")
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(data, []byte(probeIP)) != 0 {
		t.Fatal("ProbeIP not fully redacted")
	}
}
