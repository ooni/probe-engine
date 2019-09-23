package model_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/ooni/probe-engine/model"
)

type fakeTestKeys struct {
	ClientResolver string `json:"client_resolver"`
	Body           string `json:"body"`
}

func TestAddAnnotations(t *testing.T) {
	m := &model.Measurement{}
	m.AddAnnotations(map[string]string{
		"foo": "bar",
		"f":   "b",
	})
	m.AddAnnotations(map[string]string{
		"foobar": "bar",
		"f":      "b",
	})
	if len(m.Annotations) != 3 {
		t.Fatal("unexpected number of annotations")
	}
	if m.Annotations["foo"] != "bar" {
		t.Fatal("unexpected annotation")
	}
	if m.Annotations["f"] != "b" {
		t.Fatal("unexpected annotation")
	}
	if m.Annotations["foobar"] != "bar" {
		t.Fatal("unexpected annotation")
	}
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

func TestPrivacySettingsApply(t *testing.T) {
	ps := &model.PrivacySettings{}
	m := &model.Measurement{
		ProbeASN: "AS1234",
		ProbeCC:  "IT",
	}
	err := ps.Apply(m, model.LocationInfo{
		ASN:         1234,
		CountryCode: "IT",
		ProbeIP:     "8.8.8.8",
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.ProbeASN != model.DefaultProbeASNString {
		t.Fatal("ASN was not scrubbed")
	}
	if m.ProbeCC != model.DefaultProbeCC {
		t.Fatal("CC was not scrubbed")
	}
}

func TestPrivacySettingsApplyInvalidIP(t *testing.T) {
	ps := &model.PrivacySettings{}
	m := &model.Measurement{
		ProbeASN: "AS1234",
		ProbeCC:  "IT",
	}
	err := ps.Apply(m, model.LocationInfo{
		ASN:         1234,
		CountryCode: "IT",
		ProbeIP:     "", // is invalid
	})
	if err == nil {
		t.Fatal("expected an error here")
	}
}

func TestPrivacySettingsApplyMarshalError(t *testing.T) {
	ps := &model.PrivacySettings{}
	m := &model.Measurement{
		ProbeASN: "AS1234",
		ProbeCC:  "IT",
	}
	err := ps.MaybeRewriteTestKeys(
		m, "8.8.8.8", func(v interface{}) ([]byte, error) {
			return nil, errors.New("mocked error")
		})
	if err == nil {
		t.Fatal("expected an error here")
	}
}
