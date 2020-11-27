package engine

import (
	"testing"

	"github.com/ooni/probe-engine/model"
)

func TestExperimentHonoursSharingDefaults(t *testing.T) {
	measure := func(info *model.LocationInfo) *model.Measurement {
		sess := &Session{location: info}
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		exp := builder.NewExperiment()
		return exp.newMeasurement("")
	}
	type spec struct {
		name         string
		locationInfo *model.LocationInfo
		expect       func(*model.Measurement) bool
	}
	allspecs := []spec{{
		name:         "probeIP",
		locationInfo: &model.LocationInfo{ProbeIP: "8.8.8.8"},
		expect: func(m *model.Measurement) bool {
			return m.ProbeIP == model.DefaultProbeIP
		},
	}, {
		name:         "probeASN",
		locationInfo: &model.LocationInfo{ASN: 30722},
		expect: func(m *model.Measurement) bool {
			return m.ProbeASN == "AS30722"
		},
	}, {
		name:         "probeCC",
		locationInfo: &model.LocationInfo{CountryCode: "IT"},
		expect: func(m *model.Measurement) bool {
			return m.ProbeCC == "IT"
		},
	}, {
		name:         "probeNetworkName",
		locationInfo: &model.LocationInfo{NetworkName: "Vodafone Italia"},
		expect: func(m *model.Measurement) bool {
			return m.ProbeNetworkName == "Vodafone Italia"
		},
	}, {
		name:         "resolverIP",
		locationInfo: &model.LocationInfo{ResolverIP: "9.9.9.9"},
		expect: func(m *model.Measurement) bool {
			return m.ResolverIP == "9.9.9.9"
		},
	}, {
		name:         "resolverASN",
		locationInfo: &model.LocationInfo{ResolverASN: 44},
		expect: func(m *model.Measurement) bool {
			return m.ResolverASN == "AS44"
		},
	}, {
		name:         "resolverNetworkName",
		locationInfo: &model.LocationInfo{ResolverNetworkName: "Google LLC"},
		expect: func(m *model.Measurement) bool {
			return m.ResolverNetworkName == "Google LLC"
		},
	}}
	for _, spec := range allspecs {
		t.Run(spec.name, func(t *testing.T) {
			if !spec.expect(measure(spec.locationInfo)) {
				t.Fatal("expectation failed")
			}
		})
	}
}
